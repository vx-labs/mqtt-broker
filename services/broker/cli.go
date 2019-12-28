package broker

import (
	"fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
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
func (b *Broker) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryAdapter) {
	mesh.RegisterService(name, fmt.Sprintf("%s:%d", config.AdvertiseAddr, config.ServicePort))
}

func (b *Broker) Health() string {
	return "ok"
}
