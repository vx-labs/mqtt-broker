package broker

import (
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/services/broker/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (b *Broker) Serve(port int) net.Listener {
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	server := &server{
		broker: b,
	}
	pb.RegisterBrokerServiceServer(s, server)
	grpc_prometheus.Register(s)
	go s.Serve(b.listener)
	b.grpcServer = s
	return b.listener
}
func (b *Broker) Shutdown() {
	b.grpcServer.GracefulStop()
}
func (b *Broker) Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error {
	listener, err := catalog.Service("broker").ListenTCP()
	if err != nil {
		return err
	}
	b.listener = listener
	return nil
}

func (b *Broker) Health() string {
	return "ok"
}
