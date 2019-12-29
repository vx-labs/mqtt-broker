package auth

import (
	"fmt"
	"net"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/services/auth/pb"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/vx-labs/mqtt-broker/cluster"
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
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh discovery.DiscoveryAdapter) {
	err := mesh.RegisterService(name, fmt.Sprintf("%s:%d", config.AdvertiseAddr, config.ServicePort))
	if err != nil {
		panic(err)
	}
}

func (m *server) Health() string {
	return "ok"
}
