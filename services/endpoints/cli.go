package endpoints

import (
	"fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"github.com/vx-labs/mqtt-broker/network"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *api) Serve(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterDiscoveryServiceServer(s, b)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	b.grpcServer = s
	return lis
}
func (b *api) Shutdown() {
	b.grpcServer.GracefulStop()
}
func (b *api) Start(id, name string, mesh discovery.DiscoveryAdapter, catalog identity.Catalog, logger *zap.Logger) error {
	return nil
}
func (m *api) Health() string {
	return "ok"
}
