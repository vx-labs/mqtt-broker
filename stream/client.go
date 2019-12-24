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
	MaxBatchSize  int
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

func (c *streamClient) Consume(ctx context.Context, cancel chan struct{}, streamID string, f ShardConsumer, opts ...consumeOpt) {
	ticker := time.NewTicker(100 * time.Millisecond)
	heartbeatTicker := time.NewTicker(2 * time.Second)
	purgeTicker := time.NewTicker(20 * time.Second)
	rebalanceTicker := time.NewTicker(5 * time.Second)
	defaultID := uuid.New().String()
	session := streamSession{
		StreamID:     streamID,
		kv:           c.kv,
		Consumer:     f,
		ConsumerID:   defaultID,
		GroupID:      defaultID,
		MaxBatchSize: 20,
	}
	wg := sync.WaitGroup{}

	for _, opt := range opts {
		session = opt(session)
	}
	retryTick := time.NewTicker(5 * time.Second)
	for {
		err := registerConsumer(ctx, session.kv, streamID, session.GroupID, session.ConsumerID)
		if err == nil {
			break
		}
		c.logger.Error("failed to setup consumer group", zap.Error(err))
		<-retryTick.C
	}
	retryTick.Stop()
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
				err := heartbeatConsumer(ctx, session.kv, streamID, session.GroupID, session.ConsumerID)
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
			err := purgeDeadConsumers(ctx, session.kv, streamID, session.GroupID)
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
		retryTicker := time.NewTicker(5 * time.Second)
		defer retryTicker.Stop()
		for {
			consumers, err := listConsumers(ctx, session.kv, streamID, session.GroupID)
			if err != nil {
				c.logger.Error("failed to list other consumers", zap.Error(err))
				<-retryTicker.C
				continue
			}
			shards, err := c.getShards(ctx, streamID)
			if err != nil {
				c.logger.Error("failed to list shards", zap.Error(err))
				<-retryTicker.C
				continue
			}
			newAssignedShard := []string{}
			for _, shard := range getAssignedShard(shards, consumers, session.ConsumerID) {
				_, err := lockShard(ctx, session.kv, streamID, session.GroupID, shard, session.ConsumerID, 10*time.Second)
				if err != nil {
					c.logger.Warn("failed to lock shard", zap.Error(err))
					continue
				} else {
					newAssignedShard = append(newAssignedShard, shard)
				}
			}
			assignedShards = newAssignedShard
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

func (b *streamClient) consumeShard(ctx context.Context, shardId string, session streamSession) error {
	consumptionStart := time.Now()
	logger := b.logger.WithOptions(zap.Fields(zap.String("shard_id", shardId)))
	offset, err := session.getOffset(ctx, shardId)
	if err != nil {
		logger.Error("failed to get offset for shard", zap.Error(err))
		return err
	}
	next, messages, err := b.messages.GetMessages(ctx, session.StreamID, shardId, offset, session.MaxBatchSize)
	if err != nil {
		logger.Error("failed to get messages for shard", zap.Error(err))
		return err
	}
	if len(messages) == 0 {
		return nil
	}
	integrationElapsed := time.Since(consumptionStart)
	start := time.Now()
	idx, err := session.Consumer(messages)
	processorElapsed := time.Since(start)
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
		zap.Duration("shard_consumption_time", time.Since(consumptionStart)),
		zap.Duration("shard_consumption_processor_time", processorElapsed),
		zap.Duration("shard_consumption_integration_time", integrationElapsed),
	)
	return session.saveOffset(ctx, shardId, next)
}
