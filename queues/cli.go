package queues

import (
	fmt "fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/queues/pb"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	for _, lis := range b.listeners {
		lis.Close()
	}
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	err := mesh.RegisterService("queues", fmt.Sprintf("%s:%d", config.AdvertiseAddr, config.ServicePort))
	if err != nil {
		panic(err)
	}
	sessionsConn, err := mesh.DialService("sessions?tags=leader")
	if err != nil {
		panic(err)
	}
	b.sessions = sessions.NewClient(sessionsConn)
}
func (m *server) Health() string {
	return "ok"
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
