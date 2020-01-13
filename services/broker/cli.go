package broker

import (
	"fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/services/broker/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (b *Broker) Serve(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	server := &server{
		broker: b,
	}
	pb.RegisterBrokerServiceServer(s, server)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	b.grpcServer = s
	return lis
}
func (b *Broker) Shutdown() {
	b.grpcServer.GracefulStop()
}
func (b *Broker) Start(id, name string, mesh discovery.DiscoveryAdapter, catalog identity.Catalog, logger *zap.Logger) error {
	config := catalog.Get(name)
	err := mesh.RegisterTCPService(id, name, fmt.Sprintf("%s:%d", config.AdvertisedAddress(), config.AdvertisedPort()))
	if err != nil {
		panic(err)
	}
	return nil
}

func (b *Broker) Health() string {
	return "ok"
}
