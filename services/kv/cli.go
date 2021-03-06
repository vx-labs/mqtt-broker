package kv

import (
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/adapters/cp"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/services/kv/pb"

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
func (b *server) Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error {
	userService := catalog.Service(name, "rpc")
	raftService := catalog.Service(name, "cluster")
	raftRPCService := catalog.Service(name, "cluster_rpc")
	listener, err := userService.ListenTCP()
	if err != nil {
		return err
	}
	b.listener = listener
	b.state = cp.RaftSynchronizer(id, userService, raftService, raftRPCService, b, logger)
	leaderConn, err := catalog.Dial("kv", "rpc")
	if err != nil {
		panic(err)
	}
	b.leaderRPC = pb.NewKVServiceClient(leaderConn)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		var lastDeadline uint64 = 0
		for range ticker.C {
			if b.state.IsLeader() {
				now := uint64(time.Now().UnixNano())
				keys, err := b.store.ListExpiredKeys(lastDeadline, now)
				if err != nil {
					b.logger.Error("failed to list expired messages", zap.Error(err))
					continue
				}
				lastDeadline = now
				if len(keys) > 0 {
					err := b.commitEvent(&pb.KVStateTransition{
						Event: &pb.KVStateTransition_DeleteBatch{
							DeleteBatch: &pb.KVStateTransitionValueBatchDeleted{
								KeyMDs: keys,
							},
						},
					})
					if err != nil {
						b.logger.Error("failed to delete expired keys", zap.Int("expired_key_count", len(keys)), zap.Error(err))
					} else {
						b.logger.Info("deleted expired keys", zap.Int("expired_key_count", len(keys)))
					}
				}
			}
		}
	}()
	return nil
}
func (m *server) Health() (string, string) {
	if m.state == nil {
		return "critical", "state is not ready"
	}
	return m.state.Health()
}
func (m *server) Serve(port int) net.Listener {
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterKVServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(m.listener)
	m.gprcServer = s
	return m.listener
}
