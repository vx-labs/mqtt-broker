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

func (c *Client) Shutdown(ctx context.Context, id string) error {
	_, err := c.api.Shutdown(ctx, &ShutdownInput{ID: id})
	return err
}

func (c *Client) SendPublish(ctx context.Context, id string, publish *packet.Publish) error {
	_, err := c.api.SendPublish(ctx, &SendPublishInput{
		ID:      id,
		Publish: publish,
	})
	return err
}
