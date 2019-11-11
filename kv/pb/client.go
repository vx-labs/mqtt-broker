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
type deleteOpt func(KVDeleteInput) KVDeleteInput

func WithTimeToLive(ttl time.Duration) setOpt {
	return func(s KVSetInput) KVSetInput {
		s.TimeToLive = uint64(ttl.Nanoseconds())
		return s
	}
}
func (c *Client) SetWithVersion(ctx context.Context, key []byte, value []byte, version uint64, opts ...setOpt) error {
	if len(key) == 0 {
		return errors.New("invalid key")
	}
	input := KVSetInput{
		Key:        key,
		Value:      value,
		TimeToLive: 0,
		Version:    version,
	}
	for _, opt := range opts {
		input = opt(input)
	}
	_, err := c.api.Set(ctx, &input)
	return err
}

func (c *Client) Set(ctx context.Context, key []byte, value []byte, opts ...setOpt) error {
	md, err := c.GetMetadata(ctx, key)
	if err != nil {
		return err
	}
	return c.SetWithVersion(ctx, key, value, md.Version, opts...)
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
func (c *Client) GetWithMetadata(ctx context.Context, key []byte) ([]byte, *KVMetadata, error) {
	if len(key) == 0 {
		return nil, nil, errors.New("invalid key")
	}
	out, err := c.api.GetWithMetadata(ctx, &KVGetWithMetadataInput{
		Key: key,
	})
	if err != nil {
		return nil, nil, err
	}
	return out.Value, out.Metadata, nil
}
func (c *Client) GetMetadata(ctx context.Context, key []byte) (*KVMetadata, error) {
	if len(key) == 0 {
		return nil, errors.New("invalid key")
	}
	out, err := c.api.GetMetadata(ctx, &KVGetMetadataInput{
		Key: key,
	})
	if err != nil {
		return nil, err
	}
	return out.Metadata, nil
}
func (c *Client) Delete(ctx context.Context, key []byte, opts ...deleteOpt) error {
	md, err := c.GetMetadata(ctx, key)
	if err != nil {
		return err
	}
	return c.DeleteWithVersion(ctx, key, md.Version)
}
func (c *Client) DeleteWithVersion(ctx context.Context, key []byte, version uint64, opts ...deleteOpt) error {
	if len(key) == 0 {
		return errors.New("invalid key")
	}
	input := KVDeleteInput{
		Key:     key,
		Version: version,
	}
	for _, opt := range opts {
		input = opt(input)
	}
	_, err := c.api.Delete(ctx, &input)
	return err
}
