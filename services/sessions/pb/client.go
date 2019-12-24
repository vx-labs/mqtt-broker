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

type SessionFilter func(SessionFilterInput) SessionFilterInput

func ByIDFilter(id []string) SessionFilter {
	return func(s SessionFilterInput) SessionFilterInput {
		s.ID = id
		return s
	}
}

func (c *Client) All(ctx context.Context, filters ...SessionFilter) ([]*Session, error) {
	filter := SessionFilterInput{}
	for _, f := range filters {
		filter = f(filter)
	}
	out, err := c.api.All(ctx, &filter)
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Sessions), nil
}
func (c *Client) RefreshKeepAlive(ctx context.Context, id string, timestamp int64) error {
	_, err := c.api.RefreshKeepAlive(ctx, &RefreshKeepAliveInput{ID: id, Timestamp: timestamp})
	return err
}
