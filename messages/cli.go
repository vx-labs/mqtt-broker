package messages

import (
	fmt "fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/messages/pb"
	"github.com/vx-labs/mqtt-broker/network"

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
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	err := b.state.Start(name, b)
	if err != nil {
		panic(err)
	}
	leaderConn, err := mesh.DialService("messages?raft_status=leader")
	if err != nil {
		panic(err)
	}
	b.leaderRPC = pb.NewMessagesServiceClient(leaderConn)

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
