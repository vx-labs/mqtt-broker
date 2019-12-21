package pb

import (
	context "context"

	"google.golang.org/grpc"
)

type Client struct {
	api SubscriptionsServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewSubscriptionsServiceClient(conn),
	}
}
func emptyArrayIfNull(in []*Metadata) []*Metadata {
	if in == nil {
		return make([]*Metadata, 0)
	}
	return in
}
func (c *Client) Create(ctx context.Context, input SubscriptionCreateInput) error {
	_, err := c.api.Create(ctx, &input)
	return err
}
func (c *Client) Delete(ctx context.Context, id string) error {
	_, err := c.api.Delete(ctx, &SubscriptionDeleteInput{
		ID: id,
	})
	return err
}
func (c *Client) ByID(ctx context.Context, id string) (*Metadata, error) {
	return c.api.ByID(ctx, &SubscriptionByIDInput{ID: id})
}
func (c *Client) BySession(ctx context.Context, id string) ([]*Metadata, error) {
	out, err := c.api.BySession(ctx, &SubscriptionBySessionInput{SessionID: id})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Metadatas), nil
}
func (c *Client) ByTopic(ctx context.Context, tenant string, pattern []byte) ([]*Metadata, error) {
	out, err := c.api.ByTopic(ctx, &SubscriptionByTopicInput{Tenant: tenant, Topic: pattern})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Metadatas), nil
}
func (c *Client) All(ctx context.Context) ([]*Metadata, error) {
	out, err := c.api.All(ctx, &SubscriptionFilterInput{})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Metadatas), nil
}
