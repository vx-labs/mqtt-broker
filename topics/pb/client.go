package pb

import (
	context "context"

	"google.golang.org/grpc"
)

type Client struct {
	api TopicsServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewTopicsServiceClient(conn),
	}
}

func (c *Client) ByTopicPattern(ctx context.Context, tenant string, pattern []byte) ([]*RetainedMessage, error) {
	out, err := c.api.ByTopicPattern(ctx, &ByTopicPatternInput{Tenant: tenant, Pattern: pattern})
	if err != nil {
		return nil, err
	}
	return out.Messages, nil
}
