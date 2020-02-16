package endpoints

import (
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *api) Serve(port int) net.Listener {
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterDiscoveryServiceServer(s, b)
	grpc_prometheus.Register(s)
	go s.Serve(b.listener)
	b.grpcServer = s
	return b.listener
}
func (b *api) Shutdown() {
	b.grpcServer.GracefulStop()
}
func (b *api) Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error {
	listener, err := catalog.Service(name, "rpc").ListenTCP()
	if err != nil {
		return err
	}
	b.listener = listener
	return nil
}
func (m *api) Health() string {
	return "ok"
}
