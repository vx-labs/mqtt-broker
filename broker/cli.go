package broker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cenkalti/backoff"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/cluster"
	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		err = backoff.Retry(func() error {
			ctx, cancel := context.WithTimeout(b.ctx, 3*time.Second)
			defer cancel()
			_, err := b.Messages.GetStream(ctx, "messages")
			if err != nil {
				if code, ok := status.FromError(err); ok {
					if code.Code() == codes.NotFound {
						err := b.Messages.CreateStream(ctx, "messages", 3)
						if err != nil {
							b.logger.Error("failed to create stream in message store", zap.Error(err))
						}
					}
				} else {
					b.logger.Error("failed to ensure stream exists", zap.Error(err))
				}
				return err
			}
			return nil
		}, backoff.NewExponentialBackOff())
		if err != nil {
			b.logger.Fatal("failed to create stream in message store", zap.Error(err))
		}
	}()
}

func (b *Broker) Health() string {
	return "ok"
}
