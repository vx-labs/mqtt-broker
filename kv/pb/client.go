package pb

import (
	context "context"
	"errors"

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
func (c *Client) Set(ctx context.Context, key []byte, value []byte) error {
	if len(key) == 0 {
		return errors.New("invalid key")
	}
	_, err := c.api.Set(ctx, &KVSetInput{
		Key:   key,
		Value: value,
	})
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
