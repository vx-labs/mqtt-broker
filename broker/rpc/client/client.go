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

func (c *Client) ListSessions(ctx context.Context) (sessions.SessionList, error) {
	set, err := c.api.ListSessions(ctx, &rpc.SessionFilter{})
	if err != nil {
		return nil, err
	}
	return set.Sessions, nil
}
