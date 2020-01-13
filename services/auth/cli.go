package auth

import (
	"fmt"
	"net"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/services/auth/pb"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"github.com/vx-labs/mqtt-broker/network"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Serve(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterAuthServiceServer(s, b)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	b.grpcServer = s
	return lis
}
func (b *server) Shutdown() {
	b.grpcServer.GracefulStop()
}
func (b *server) Start(id, name string, mesh discovery.DiscoveryAdapter, catalog identity.Catalog, logger *zap.Logger) error {
	config := catalog.Get(name)
	err := mesh.RegisterTCPService(id, name, fmt.Sprintf("%s:%d", config.AdvertisedAddress(), config.AdvertisedPort()))
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *server) Health() string {
	return "ok"
}
