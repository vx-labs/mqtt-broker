package pb

import (
	context "context"
	"errors"

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
func (c *Client) CreateStream(ctx context.Context, streamId string, shardCount int) error {
	if len(streamId) == 0 {
		return errors.New("invalid stream id")
	}
	_, err := c.api.CreateStream(ctx, &MessageCreateStreamInput{
		ID:         streamId,
		ShardCount: int64(shardCount),
	})
	return err
}
func (c *Client) GetStream(ctx context.Context, streamId string) (*StreamConfig, error) {
	if len(streamId) == 0 {
		return nil, errors.New("invalid stream id")
	}
	out, err := c.api.GetStream(ctx, &MessageGetStreamInput{
		ID: streamId,
	})
	if err != nil {
		return nil, err
	}
	return out.Config, nil
}
func (c *Client) GetStreamStatistics(ctx context.Context, streamId string) (*StreamStatistics, error) {
	if len(streamId) == 0 {
		return nil, errors.New("invalid stream id")
	}
	out, err := c.api.GetStreamStatistics(ctx, &MessageGetStreamStatisticsInput{
		ID: streamId,
	})
	if err != nil {
		return nil, err
	}
	return out.Statistics, nil
}
func (c *Client) ListStreams(ctx context.Context) ([]*StreamConfig, error) {
	out, err := c.api.ListStreams(ctx, &MessageListStreamInput{})
	if err != nil {
		return nil, err
	}
	return out.Streams, nil
}
func (c *Client) Put(ctx context.Context, streamId string, shardKey string, payload []byte) error {
	if len(streamId) == 0 {
		return errors.New("invalid stream id")
	}
	if len(shardKey) == 0 {
		return errors.New("invalid shard key")
	}
	_, err := c.api.PutMessage(ctx, &MessagePutMessageInput{
		StreamID: streamId,
		Payload:  payload,
		ShardKey: shardKey,
	})
	return err
}

type MessageBatch struct {
	StreamID string
	ShardKey string
	Payload  []byte
}

func (c *Client) PutBatch(ctx context.Context, payload []MessageBatch) error {
	batches := make([]*MessagePutMessageInput, len(payload))
	for idx := range payload {
		batches[idx] = &MessagePutMessageInput{
			StreamID: payload[idx].StreamID,
			ShardKey: payload[idx].ShardKey,
			Payload:  payload[idx].Payload,
		}
	}
	_, err := c.api.PutMessageBatch(ctx, &MessagePutMessageBatchInput{
		Batches: batches,
	})
	return err
}
func (c *Client) GetMessages(ctx context.Context, streamId string, shardId string, fromOffset uint64, count int) (uint64, []*StoredMessage, error) {
	out, err := c.api.GetMessages(ctx, &MessageGetMessagesInput{
		StreamID: streamId,
		ShardID:  shardId,
		Offset:   fromOffset,
		MaxCount: uint64(count),
	})
	if err != nil {
		return 0, nil, err
	}
	return out.NextOffset, out.Messages, nil
}
