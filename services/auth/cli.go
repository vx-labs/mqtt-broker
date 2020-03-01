package auth

import (
	"net"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/services/auth/pb"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/vx-labs/mqtt-broker/network"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Serve(port int) net.Listener {
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterAuthServiceServer(s, b)
	grpc_prometheus.Register(s)
	go s.Serve(b.listener)
	b.grpcServer = s
	return b.listener
}
func (b *server) Shutdown() {
	b.grpcServer.GracefulStop()
}
func (b *server) Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error {
	listener, err := catalog.Service(name, "rpc").ListenTCP()
	if err != nil {
		return err
	}
	b.listener = listener
	return nil
}

func (m *server) Health() (string, string) {
	return "ok", ""
}
