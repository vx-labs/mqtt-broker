package queues

import (
	fmt "fmt"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/services/queues/pb"
	sessions "github.com/vx-labs/mqtt-broker/services/sessions/pb"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	err := b.state.Shutdown()
	if err != nil {
		b.logger.Error("failed to shutdown raft instance cleanly", zap.Error(err))
	}
	b.gprcServer.GracefulStop()
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	var err error
	sessionConn, err := mesh.DialService("sessions")
	if err != nil {
		logger.Fatal("failed to dial session service")
	}
	b.sessions = sessions.NewClient(sessionConn)
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	err = b.state.Start(name, b)
	if err != nil {
		panic(err)
	}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			if b.state.IsLeader() {
				expiredInflights := b.store.GetExpiredInflights(time.Now())
				if len(expiredInflights) > 0 {
					err := b.commitEvent(&pb.QueuesStateTransition{
						Kind: MessageInflightExpired,
						MessageInflightExpired: &pb.QueueStateTransitionMessageInflightExpired{
							ExpiredInflights: expiredInflights,
						},
					})
					if err != nil {
						b.logger.Error("failed to commit expired inflight message", zap.Error(err))
					}
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
	m.gprcServer = s
	return lis
}
