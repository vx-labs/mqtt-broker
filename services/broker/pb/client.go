package pb

import (
	"context"

	"github.com/vx-labs/mqtt-broker/transport"

	"github.com/vx-labs/mqtt-protocol/packet"
	"google.golang.org/grpc"
)

type Client struct {
	api BrokerServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewBrokerServiceClient(conn),
	}
}

func (c *Client) CloseSession(ctx context.Context, id string) error {
	_, err := c.api.CloseSession(ctx, &CloseSessionInput{ID: id})
	return err
}

func (c *Client) Connect(ctx context.Context, metadata transport.Metadata, connect *packet.Connect) (string, string, string, *packet.ConnAck, error) {
	out, err := c.api.Connect(ctx, &ConnectInput{Connect: connect, TransportMetadata: &TransportMetadata{
		Encrypted:     metadata.Encrypted,
		Name:          metadata.Name,
		RemoteAddress: metadata.RemoteAddress,
		Endpoint:      metadata.Endpoint,
	}})
	if out == nil {
		return "", "", "", &packet.ConnAck{
			Header:     &packet.Header{},
			ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
		}, err
	}
	return out.ID, out.Token, out.RefreshToken, out.ConnAck, err
}
func (c *Client) Publish(ctx context.Context, id string, publish *packet.Publish) (*packet.PubAck, error) {
	out, err := c.api.Publish(ctx, &PublishInput{ID: id, Publish: publish})
	if out == nil {
		return nil, err
	}
	return out.PubAck, err
}
func (c *Client) Subscribe(ctx context.Context, id string, subscribe *packet.Subscribe) (*packet.SubAck, error) {
	out, err := c.api.Subscribe(ctx, &SubscribeInput{ID: id, Subscribe: subscribe})
	if out == nil {
		return nil, err
	}
	return out.SubAck, err
}
func (c *Client) Unsubscribe(ctx context.Context, id string, unsubscribe *packet.Unsubscribe) (*packet.UnsubAck, error) {
	out, err := c.api.Unsubscribe(ctx, &UnsubscribeInput{ID: id, Unsubscribe: unsubscribe})
	if out == nil {
		return nil, err
	}
	return out.UnsubAck, err
}
func (c *Client) Disconnect(ctx context.Context, id string, disconnect *packet.Disconnect) error {
	_, err := c.api.Disconnect(ctx, &DisconnectInput{ID: id, Disconnect: disconnect})
	return err
}
func (c *Client) PingReq(ctx context.Context, id string, pingreq *packet.PingReq) (string, *packet.PingResp, error) {
	out, err := c.api.PingReq(ctx, &PingReqInput{ID: id, PingReq: pingreq})
	if out == nil {
		return "", nil, err
	}
	return out.RefreshedToken, out.PingResp, err
}
