package pb

import (
	context "context"

	packet "github.com/vx-labs/mqtt-protocol/packet"
	"google.golang.org/grpc"
)

type Client struct {
	api MessagesServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewMessagesServiceClient(conn),
	}
}
func (c *Client) Put(ctx context.Context, tenant string, publish *packet.Publish) error {
	_, err := c.api.PutMessage(ctx, &MessagePutMessageInput{
		Tenant:  tenant,
		Publish: publish,
	})
	return err
}

type MessageBatch struct {
	Tenant  string
	Publish *packet.Publish
}

func (c *Client) PutBatch(ctx context.Context, payload []MessageBatch) error {
	batches := make([]*MessagePutMessageInput, len(payload))
	for idx := range payload {
		batches[idx] = &MessagePutMessageInput{
			Tenant:  payload[idx].Tenant,
			Publish: payload[idx].Publish,
		}
	}
	_, err := c.api.PutMessageBatch(ctx, &MessagePutMessageBatchInput{
		Batches: batches,
	})
	return err
}
