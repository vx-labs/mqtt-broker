package client

import (
	"context"

	"github.com/vx-labs/mqtt-broker/transport"

	"github.com/vx-labs/mqtt-broker/broker/rpc"
	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-protocol/packet"
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
			Transport: nil,
			Metadata:  *out.Sessions[idx],
		}
	}
	return set, nil
}

func (c *Client) Connect(ctx context.Context, metadata transport.Metadata, connect *packet.Connect) (string, *packet.ConnAck, error) {
	out, err := c.api.Connect(ctx, &rpc.ConnectInput{Connect: connect, TransportMetadata: &rpc.TransportMetadata{
		Encrypted:     metadata.Encrypted,
		Name:          metadata.Name,
		RemoteAddress: metadata.RemoteAddress,
	}})
	if err != nil {
		return "", &packet.ConnAck{
			Header:     &packet.Header{},
			ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
		}, err
	}
	return out.ID, out.ConnAck, err
}
func (c *Client) Publish(ctx context.Context, id string, publish *packet.Publish) (*packet.PubAck, error) {
	out, err := c.api.Publish(ctx, &rpc.PublishInput{ID: id, Publish: publish})
	return out.PubAck, err
}
func (c *Client) Subscribe(ctx context.Context, id string, subscribe *packet.Subscribe) (*packet.SubAck, error) {
	out, err := c.api.Subscribe(ctx, &rpc.SubscribeInput{ID: id, Subscribe: subscribe})
	return out.SubAck, err
}
func (c *Client) Unsubscribe(ctx context.Context, id string, unsubscribe *packet.Unsubscribe) (*packet.UnsubAck, error) {
	out, err := c.api.Unsubscribe(ctx, &rpc.UnsubscribeInput{ID: id, Unsubscribe: unsubscribe})
	return out.UnsubAck, err
}
func (c *Client) Disconnect(ctx context.Context, id string, disconnect *packet.Disconnect) error {
	_, err := c.api.Disconnect(ctx, &rpc.DisconnectInput{ID: id, Disconnect: disconnect})
	return err
}
