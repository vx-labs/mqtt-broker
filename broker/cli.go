package broker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/vx-labs/mqtt-broker/cluster"
	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (b *Broker) Serve(port int) net.Listener {
	return Serve(port, b)
}
func (b *Broker) Shutdown() {
	b.Stop()
}
func (b *Broker) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	mesh.RegisterService(name, fmt.Sprintf("%s:%d", config.AdvertiseAddr, config.ServicePort))
	go func() {
		messagesConn, err := mesh.DialService("messages?tags=leader")
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
						err := b.Messages.CreateStream(ctx, "messages", 1)
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
