package broker

import (
	"fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/cluster"
	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	"github.com/vx-labs/mqtt-broker/network"
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
func (b *Broker) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	mesh.RegisterService(name, fmt.Sprintf("%s:%d", config.AdvertiseAddr, config.ServicePort))
	go func() {
		messagesConn, err := mesh.DialService("messages?raft_status=leader")
		if err != nil {
			panic(err)
		}
		b.Messages = messages.NewClient(messagesConn)
	}()
}

func (b *Broker) Health() string {
	return "ok"
}
