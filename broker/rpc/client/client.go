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
func (c *Client) ListSessions(ctx context.Context) (sessions.SessionSet, error) {
	out, err := c.api.ListSessions(ctx, &rpc.SessionFilter{})
	if err != nil {
		return sessions.SessionSet{}, err
	}
	set := make(sessions.SessionSet, len(out.Sessions))
	for idx := range set {
		set[idx] = sessions.Session{
			Close: func() error {
				return c.CloseSession(ctx, out.Sessions[idx].ID)
			},
			SessionMD: *out.Sessions[idx],
		}
	}
	return set, nil
}
