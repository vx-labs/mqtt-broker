package client

import (
	"context"

	"github.com/vx-labs/mqtt-broker/broker/rpc"
	"github.com/vx-labs/mqtt-broker/sessions"
	"google.golang.org/grpc"
)

type Client struct {
	api rpc.BrokerServiceClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		api: rpc.NewBrokerServiceClient(conn),
	}
}

func (c *Client) CloseSession(ctx context.Context, id string) error {
	_, err := c.api.CloseSession(ctx, &rpc.CloseSessionInput{ID: id})
	return err
}
func (c *Client) ListSessions(ctx context.Context) (sessions.SessionList, error) {
	set, err := c.api.ListSessions(ctx, &rpc.SessionFilter{})
	if err != nil {
		return sessions.SessionList{}, err
	}
	return sessions.SessionList{
		Sessions: set.Sessions,
	}, nil
}
