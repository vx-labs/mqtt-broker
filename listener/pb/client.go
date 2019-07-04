package pb

import (
	"context"

	"github.com/vx-labs/mqtt-protocol/packet"
	"google.golang.org/grpc"
)

type Client struct {
	api ListenerServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewListenerServiceClient(conn),
	}
}

func (c *Client) CloseSession(ctx context.Context, id string) error {
	_, err := c.api.CloseSession(ctx, &CloseSessionInput{ID: id})
	return err
}
func (c *Client) Close(ctx context.Context) error {
	_, err := c.api.Close(ctx, &CloseInput{})
	return err
}
func (c *Client) Publish(ctx context.Context, id string, publish *packet.Publish) error {
	_, err := c.api.Publish(ctx, &PublishInput{
		ID:      id,
		Publish: publish,
	})
	return err
}
