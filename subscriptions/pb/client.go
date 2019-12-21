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
func emptyArrayIfNull(in []*Subscription) []*Subscription {
	if in == nil {
		return make([]*Subscription, 0)
	}
	return in
}
func (c *Client) ByID(ctx context.Context, id string) (*Subscription, error) {
	return c.api.ByID(ctx, &SubscriptionByIDInput{ID: id})
}
func (c *Client) BySession(ctx context.Context, id string) ([]*Subscription, error) {
	out, err := c.api.BySession(ctx, &SubscriptionBySessionInput{SessionID: id})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Subscriptions), nil
}
func (c *Client) ByTopic(ctx context.Context, tenant string, pattern []byte) ([]*Subscription, error) {
	out, err := c.api.ByTopic(ctx, &SubscriptionByTopicInput{Tenant: tenant, Topic: pattern})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Subscriptions), nil
}
func (c *Client) All(ctx context.Context) ([]*Subscription, error) {
	out, err := c.api.All(ctx, &SubscriptionFilterInput{})
	if err != nil {
		return nil, err
	}
	return emptyArrayIfNull(out.Subscriptions), nil
}
