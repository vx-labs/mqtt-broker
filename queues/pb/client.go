package pb

import (
	context "context"

	packet "github.com/vx-labs/mqtt-protocol/packet"
	"google.golang.org/grpc"
)

type Client struct {
	api QueuesServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewQueuesServiceClient(conn),
	}
}
func (c *Client) Create(ctx context.Context, id string) error {
	_, err := c.api.Create(ctx, &QueueCreateInput{Id: id})
	return err
}
func (c *Client) Delete(ctx context.Context, id string) error {
	_, err := c.api.Delete(ctx, &QueueDeleteInput{
		Id: id,
	})
	return err
}
func (c *Client) PutMessage(ctx context.Context, id string, publish *packet.Publish) error {
	_, err := c.api.PutMessage(ctx, &QueuePutMessageInput{
		Id:      id,
		Publish: publish,
	})
	return err
}

type MessageBatch struct {
	ID      string
	Publish *packet.Publish
}

func (c *Client) PutMessageBatch(ctx context.Context, payload []MessageBatch) error {
	batches := make([]*QueuePutMessageInput, len(payload))
	for idx := range payload {
		batches[idx] = &QueuePutMessageInput{
			Id:      payload[idx].ID,
			Publish: payload[idx].Publish,
		}
	}
	_, err := c.api.PutMessageBatch(ctx, &QueuePutMessageBatchInput{
		Batches: batches,
	})
	return err
}
func (c *Client) GetMessages(ctx context.Context, id string, offset uint64) (uint64, []*packet.Publish, error) {
	resp, err := c.api.GetMessages(ctx, &QueueGetMessagesInput{
		Id:     id,
		Offset: offset,
	})
	if err != nil {
		return 0, nil, err
	}
	return resp.Offset, resp.Publishes, nil
}
func (c *Client) GetMessagesBatch(ctx context.Context, input []*QueueGetMessagesInput) ([]*QueueGetMessagesOutput, error) {
	resp, err := c.api.GetMessagesBatch(ctx, &QueueGetMessagesBatchInput{
		Batches: input,
	})
	if err != nil {
		return nil, err
	}
	return resp.Batches, nil
}
