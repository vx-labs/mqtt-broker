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
func (c *Client) AckMessage(ctx context.Context, id string, ackOffset uint64) error {
	_, err := c.api.AckMessage(ctx, &AckMessageInput{
		Id:     id,
		Offset: ackOffset,
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
func (c *Client) StreamMessages(ctx context.Context, id string, f func(uint64, *packet.Publish) error) error {
	stream, err := c.api.StreamMessages(ctx, &QueueGetMessagesInput{
		Id:     id,
		Offset: 0,
	})
	if err != nil {
		return err
	}
	for {
		message, err := stream.Recv()
		if err != nil {
			return err
		}
		for _, batch := range message.Batches {
			err = f(batch.AckOffset, batch.Publish)
			if err != nil {
				return err
			}
		}
	}
}
func (c *Client) ListQueues(ctx context.Context) ([]string, error) {
	out, err := c.api.List(ctx, &QueuesListInput{})
	if err != nil {
		return nil, err
	}
	return out.QueueIDs, nil
}
func (c *Client) GetQueueStatistics(ctx context.Context, queueID string) (*QueueStatistics, error) {
	out, err := c.api.GetStatistics(ctx, &QueueGetStatisticsInput{
		ID: queueID,
	})
	if err != nil {
		return nil, err
	}
	return out.Statistics, nil
}
