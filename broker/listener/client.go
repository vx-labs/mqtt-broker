package listener

import (
	"context"

	"github.com/vx-labs/mqtt-broker/broker/listener/pb"
	"github.com/vx-labs/mqtt-protocol/packet"
	"google.golang.org/grpc"
)

type Client struct {
	api pb.ListenerServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: pb.NewListenerServiceClient(conn),
	}
}

func (c *Client) CloseSession(ctx context.Context, id string) error {
	_, err := c.api.CloseSession(ctx, &pb.CloseSessionInput{ID: id})
	return err
}
func (c *Client) Close(ctx context.Context) error {
	_, err := c.api.Close(ctx, &pb.CloseInput{})
	return err
}
func (c *Client) Publish(ctx context.Context, id string, publish *packet.Publish) error {
	_, err := c.api.Publish(ctx, &pb.PublishInput{
		ID:      id,
		Publish: publish,
	})
	return err
}
