package stream

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	kv "github.com/vx-labs/mqtt-broker/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type streamClient struct {
	kv       *kv.Client
	messages *messages.Client
	logger   *zap.Logger
}

func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

type ShardConsumer func([]*messages.StoredMessage) (int, error)

func NewClient(kvClient *kv.Client, messagesClient *messages.Client, logger *zap.Logger) *streamClient {
	return &streamClient{
		kv:       kvClient,
		messages: messagesClient,
		logger:   logger,
	}
}

type streamSession struct {
	kv            *kv.Client
	ConsumerID    string
	GroupID       string
	StreamID      string
	Consumer      ShardConsumer
	DefaultOffset uint64
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

func (c *streamClient) getShards(ctx context.Context, streamID string) ([]string, error) {
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

func (c *streamClient) Consume(ctx context.Context, cancel chan struct{}, streamID string, f ShardConsumer, opts ...consumeOpt) {
	ticker := time.NewTicker(200 * time.Millisecond)
	heartbeatTicker := time.NewTicker(2 * time.Second)
	purgeTicker := time.NewTicker(20 * time.Second)
	rebalanceTicker := time.NewTicker(5 * time.Second)
	defaultID := uuid.New().String()
	session := streamSession{
		StreamID:   streamID,
		kv:         c.kv,
		Consumer:   f,
		ConsumerID: defaultID,
		GroupID:    defaultID,
	}
	wg := sync.WaitGroup{}

	for _, opt := range opts {
		session = opt(session)
	}
	err := registerConsumer(ctx, session.kv, streamID, session.GroupID, session.ConsumerID)
	if err != nil {
		panic("failed to setup consumer group")

	}
	defer unregisterConsumer(ctx, session.kv, streamID, session.GroupID, session.ConsumerID)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer heartbeatTicker.Stop()
		for {
			select {
			case <-cancel:
				return
			case <-ctx.Done():
				return
			case <-heartbeatTicker.C:
				err = heartbeatConsumer(ctx, session.kv, streamID, session.GroupID, session.ConsumerID)
				if err != nil {
					c.logger.Error("failed to heartbeat group", zap.Error(err))
				}
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer purgeTicker.Stop()

		for {
			err = purgeDeadConsumers(ctx, session.kv, streamID, session.GroupID)
			if err != nil {
				c.logger.Error("failed to purge dead consumers", zap.Error(err))
			}
			select {
			case <-cancel:
				return
			case <-ctx.Done():
				return
			case <-purgeTicker.C:
			}
		}
	}()

	assignedShards := []string{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer rebalanceTicker.Stop()
		for {
			consumers, err := listConsumers(ctx, session.kv, streamID, session.GroupID)
			if err != nil {
				c.logger.Error("failed to list other consumers", zap.Error(err))
				continue
			}
			shards, err := c.getShards(ctx, streamID)
			if err != nil {
				c.logger.Error("failed to list shards", zap.Error(err))
				continue
			}
			assignedShards = getAssignedShard(shards, consumers, session.ConsumerID)
			select {
			case <-cancel:
				return
			case <-ctx.Done():
				return
			case <-rebalanceTicker.C:
			}
		}
	}()
	defer ticker.Stop()
	for {
		select {
		case <-cancel:
			wg.Wait()
			return
		case <-ctx.Done():
			wg.Wait()
			return
		case <-ticker.C:
			for _, shard := range assignedShards {
				err := c.consumeShard(ctx, shard, session)
				if err != nil {
					c.logger.Error("failed to consume shard", zap.Error(err))
				}
			}
		}
	}
}

type shardSession struct {
	cancel      chan struct{}
	shardID     string
	session     streamSession
	lockVersion uint64
	messages    *messages.Client
}

func newShardSession(shardID string, session streamSession) *shardSession {
	return &shardSession{
		cancel:  make(chan struct{}),
		shardID: shardID,
		session: session,
	}
}

func (shardSession *shardSession) Consume(ctx context.Context, logger *zap.Logger) error {
	lockVersion, err := lockShard(ctx, shardSession.session.kv, shardSession.session.StreamID, shardSession.session.GroupID, shardSession.shardID, shardSession.session.ConsumerID, 20*time.Second)
	if err != nil {
		return errors.New("unable to lock shard")
	}
	defer unlockShard(ctx, shardSession.session.kv, shardSession.session.StreamID, shardSession.session.GroupID, shardSession.shardID, shardSession.session.ConsumerID)

	logger = logger.WithOptions(zap.Fields(zap.String("shard_id", shardSession.shardID)))

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		case <-shardSession.cancel:
			return nil
		}
		err := renewShardlock(ctx, shardSession.session.kv, shardSession.session.StreamID, shardSession.session.GroupID, shardSession.shardID, shardSession.session.ConsumerID, 20*time.Second, lockVersion)
		if err != nil {
			logger.Info("unable to renew lock shard")
			return err
		}
		err = shardSession.run(ctx, logger)
		if err != nil {
			logger.Error("failed to consume messages in shard", zap.Error(err))
		}
	}
}
func (shardSession *shardSession) run(ctx context.Context, logger *zap.Logger) error {
	offset, err := shardSession.session.getOffset(ctx, shardSession.shardID)
	if err != nil {
		logger.Error("failed to get offset for shard", zap.Error(err))
		return err
	}
	next, messages, err := shardSession.messages.GetMessages(ctx, shardSession.session.StreamID, shardSession.shardID, offset, 20)
	if err != nil {
		logger.Error("failed to get messages for shard", zap.Error(err))
		return err
	}
	if len(messages) == 0 {
		return nil
	}
	start := time.Now()
	idx, err := shardSession.session.Consumer(messages)
	if err != nil {
		logger.Error("failed to consume shard", zap.Uint64("shard_offset", messages[idx].Offset), zap.Error(err))
		err = shardSession.session.saveOffset(ctx, shardSession.shardID, offset)
		if err != nil {
			logger.Error("failed to save offset", zap.Uint64("shard_offset", messages[idx].Offset), zap.Error(err))
			return err
		}
		return err
	}
	iteratorAge := time.Since(time.Unix(0, int64(next)))
	logger.Info("shard messages consumed",
		zap.Uint64("shard_offset", offset),
		zap.Int("shard_message_count", len(messages)),
		zap.Duration("shard_iterator_age", iteratorAge),
		zap.Duration("shard_consumption_time", time.Since(start)),
	)
	return shardSession.session.saveOffset(ctx, shardSession.shardID, next)
}

func (b *streamClient) consumeShard(ctx context.Context, shardId string, session streamSession) error {
	_, err := lockShard(ctx, b.kv, session.StreamID, session.GroupID, shardId, session.ConsumerID, 20*time.Second)
	if err != nil {
		return errors.New("unable to lock shard")
	}
	defer unlockShard(ctx, b.kv, session.StreamID, session.GroupID, shardId, session.ConsumerID)

	logger := b.logger.WithOptions(zap.Fields(zap.String("shard_id", shardId)))
	offset, err := session.getOffset(ctx, shardId)
	if err != nil {
		logger.Error("failed to get offset for shard", zap.Error(err))
		return err
	}
	next, messages, err := b.messages.GetMessages(ctx, session.StreamID, shardId, offset, 20)
	if err != nil {
		logger.Error("failed to get messages for shard", zap.Error(err))
		return err
	}
	if len(messages) == 0 {
		return nil
	}
	start := time.Now()
	idx, err := session.Consumer(messages)
	if err != nil {
		logger.Error("failed to consume shard", zap.Uint64("shard_offset", messages[idx].Offset), zap.Error(err))
		err = session.saveOffset(ctx, shardId, offset)
		if err != nil {
			logger.Error("failed to save offset", zap.Uint64("shard_offset", messages[idx].Offset), zap.Error(err))
			return err
		}
		return err
	}
	iteratorAge := time.Since(time.Unix(0, int64(next)))
	b.logger.Info("shard messages consumed",
		zap.Uint64("shard_offset", offset),
		zap.Int("shard_message_count", len(messages)),
		zap.Duration("shard_iterator_age", iteratorAge),
		zap.Duration("shard_consumption_time", time.Since(start)),
	)
	return session.saveOffset(ctx, shardId, next)
}
