package pb

import (
	context "context"
	"errors"
	"time"

	"google.golang.org/grpc"
)

type Client struct {
	api KVServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewKVServiceClient(conn),
	}
}

type setOpt func(KVSetInput) KVSetInput

func WithTimeToLive(ttl time.Duration) setOpt {
	return func(s KVSetInput) KVSetInput {
		s.TimeToLive = uint64(ttl.Nanoseconds())
		return s
	}
}

func (c *Client) Set(ctx context.Context, key []byte, value []byte, opts ...setOpt) error {
	if len(key) == 0 {
		return errors.New("invalid key")
	}
	input := KVSetInput{
		Key:        key,
		Value:      value,
		TimeToLive: 0,
	}
	for _, opt := range opts {
		input = opt(input)
	}
	_, err := c.api.Set(ctx, &input)
	return err
}
func (c *Client) Get(ctx context.Context, key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("invalid key")
	}
	out, err := c.api.Get(ctx, &KVGetInput{
		Key: key,
	})
	if err != nil {
		return nil, err
	}
	return out.Value, nil
}
func (c *Client) Delete(ctx context.Context, key []byte) error {
	if len(key) == 0 {
		return errors.New("invalid key")
	}
	_, err := c.api.Delete(ctx, &KVDeleteInput{
		Key: key,
	})
	return err
}
