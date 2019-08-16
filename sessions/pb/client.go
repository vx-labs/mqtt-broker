package pb

import (
	context "context"

	"google.golang.org/grpc"
)

type Client struct {
	api SessionsServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewSessionsServiceClient(conn),
	}
}
func emptyArrayIfNull(in []*Session) []*Session {
	if in == nil {
		return make([]*Session, 0)
	}
	return in
}
func (c *Client) Create(ctx context.Context, input SessionCreateInput) error {
	_, err := c.api.Create(ctx, &input)
	return err
}
func (c *Client) Delete(ctx context.Context, id string) error {
	_, err := c.api.Delete(ctx, &SessionDeleteInput{
		ID: id,
	})
	return err
}
func (c *Client) ByID(ctx context.Context, id string) (*Session, error) {
	return c.api.ByID(ctx, &SessionByIDInput{ID: id})
}
func (c *Client) ByClientID(ctx context.Context, id string) ([]*Session, error) {
	out, err := c.api.ByClientID(ctx, &SessionByClientIDInput{ClientID: id})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Sessions), nil
}
func (c *Client) ByPeer(ctx context.Context, id string) ([]*Session, error) {
	out, err := c.api.ByPeer(ctx, &SessionByPeerInput{Peer: id})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Sessions), nil
}
func (c *Client) All(ctx context.Context) ([]*Session, error) {
	out, err := c.api.All(ctx, &SessionFilterInput{})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Sessions), nil
}
