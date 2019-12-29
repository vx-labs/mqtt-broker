package messages

import (
	"context"
	fmt "fmt"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/adapters/cp"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/services/messages/pb"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	err := b.state.Shutdown()
	if err != nil {
		b.logger.Error("failed to shutdown raft state", zap.Error(err))
	}
	b.gprcServer.GracefulStop()
	b.store.Close()
}
func (b *server) Start(id, name string, mesh discovery.DiscoveryAdapter, catalog identity.Catalog, logger *zap.Logger) error {
	userService := discovery.NewServiceFromIdentity(catalog.Get(name), mesh)
	raftService := discovery.NewServiceFromIdentity(catalog.Get(fmt.Sprintf("%s_gossip", name)), mesh)
	raftRPCService := discovery.NewServiceFromIdentity(catalog.Get(fmt.Sprintf("%s_gossip_rpc", name)), mesh)
	err := userService.Register()
	if err != nil {
		logger.Error("failed to register service")
		return err
	}

	b.state = cp.RaftSynchronizer(id, userService, raftService, raftRPCService, b, logger)
	leaderConn, err := mesh.DialService("messages?raft_status=leader")
	if err != nil {
		panic(err)
	}
	b.leaderRPC = pb.NewMessagesServiceClient(leaderConn)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if b.state.IsLeader() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				for _, streamConfig := range b.config.InitialsStreams {
					if !b.store.Exists(streamConfig.ID) {
						_, err := b.CreateStream(ctx, &pb.MessageCreateStreamInput{
							ID:         streamConfig.ID,
							ShardCount: streamConfig.ShardCount,
						})
						if err != nil {
							logger.Error("failed to create initial streams", zap.Error(err))
						}
					}
				}
				return
			}
		}
	}()
	return nil
}
func (m *server) Health() string {
	return m.state.Health()
}
func (m *server) Serve(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterMessagesServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	return lis
}
