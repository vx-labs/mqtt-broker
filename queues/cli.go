package queues

import (
	fmt "fmt"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/queues/pb"
	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	for _, lis := range b.listeners {
		lis.Close()
	}
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	var err error
	sessionConn, err := mesh.DialService("sessions?raft_status=leader")
	if err != nil {
		logger.Fatal("failed to dial session service")
	}
	b.sessions = sessions.NewClient(sessionConn)
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	go func() {
		err := b.state.Start(name, b)
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			if b.state.IsLeader() {
				b.gcExpiredQueues()
				err := b.store.TickInflights(time.Now())
				if err != nil {
					b.logger.Error("failed to expire inflight messages", zap.Error(err))
				}
			}
		}
	}()
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
	pb.RegisterQueuesServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	return lis
}
