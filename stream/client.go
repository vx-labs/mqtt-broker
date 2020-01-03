package stream

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	kv "github.com/vx-labs/mqtt-broker/services/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	kv       *kv.Client
	messages *messages.Client
	logger   *zap.Logger
	sessions []*streamSession
	mtx      sync.Mutex
}

func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

type ShardConsumer func([]*messages.StoredMessage) (int, error)

func NewClient(kvClient *kv.Client, messagesClient *messages.Client, logger *zap.Logger) *Client {
	return &Client{
		kv:       kvClient,
		messages: messagesClient,
		logger:   logger,
	}
}

type streamSession struct {
	kv            *kv.Client
	messages      *messages.Client
	ConsumerID    string
	GroupID       string
	StreamID      string
	Consumer      ShardConsumer
	MaxBatchSize  int
	DefaultOffset uint64
	logger        *zap.Logger
	cancel        chan struct{}
	done          chan struct{}
}

func (c *streamSession) resetOffset(ctx context.Context, shardID string) error {
	return c.kv.Delete(ctx, c.offsetKey(shardID))
}
func (c *streamSession) offsetKey(shardID string) []byte {
	return []byte(fmt.Sprintf("stream/%s/%s/%s/offset", c.StreamID, shardID, c.GroupID))
}
func (c *streamSession) saveOffset(ctx context.Context, shardID string, offset uint64) error {
	return c.kv.Set(ctx, c.offsetKey(shardID), uint64ToBytes(offset))
}
func (c *streamSession) getOffset(ctx context.Context, shardID string) (uint64, error) {
	out, err := c.kv.Get(ctx, c.offsetKey(shardID))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return c.DefaultOffset, nil
			}
		}
		return 0, err
	}
	return binary.BigEndian.Uint64(out), nil
}

func (c *Client) getShards(ctx context.Context, streamID string) ([]string, error) {
	streamConfig, err := c.messages.GetStream(ctx, streamID)
	if err != nil {
		return nil, err
	}
	return streamConfig.ShardIDs, nil
}

type consumeOpt func(streamSession) streamSession

func WithConsumerGroupID(id string) consumeOpt {
	return func(s streamSession) streamSession {
		s.GroupID = id
		return s
	}
}

func WithConsumerID(id string) consumeOpt {
	return func(s streamSession) streamSession {
		s.ConsumerID = id
		return s
	}
}
func WithMaxBatchSize(size int) consumeOpt {
	return func(s streamSession) streamSession {
		s.MaxBatchSize = size
		return s
	}
}

type offsetBehaviour byte

const (
	OFFSET_BEHAVIOUR_FROM_START offsetBehaviour = iota
	OFFSET_BEHAVIOUR_FROM_NOW
)

func WithInitialOffsetBehaviour(b offsetBehaviour) consumeOpt {
	return func(s streamSession) streamSession {
		switch b {
		case OFFSET_BEHAVIOUR_FROM_NOW:
			s.DefaultOffset = uint64(time.Now().UnixNano())
		case OFFSET_BEHAVIOUR_FROM_START:
			s.DefaultOffset = 0
		}
		return s
	}
}

func (c *Client) Shutdown() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for _, session := range c.sessions {
		session.Shutdown()
	}
}
func (c *Client) ConsumeStream(ctx context.Context, streamID string, f ShardConsumer, opts ...consumeOpt) {
	defaultID := uuid.New().String()
	session := streamSession{
		StreamID:     streamID,
		kv:           c.kv,
		messages:     c.messages,
		logger:       c.logger,
		Consumer:     f,
		ConsumerID:   defaultID,
		GroupID:      defaultID,
		MaxBatchSize: 20,
		cancel:       make(chan struct{}),
		done:         make(chan struct{}),
	}
	for _, opt := range opts {
		session = opt(session)
	}
	c.mtx.Lock()
	c.sessions = append(c.sessions, &session)
	c.mtx.Unlock()
	session.Start(ctx, streamID, f)
}
