package stream

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

func (session *streamSession) register(ctx context.Context, streamID string) {
	retryTick := time.NewTicker(5 * time.Second)
	defer retryTick.Stop()
	for {
		err := registerConsumer(ctx, session.kv, streamID, session.GroupID, session.ConsumerID)
		if err == nil {
			return
		}
		session.logger.Error("failed to setup consumer group", zap.Error(err))
		select {
		case <-session.cancel:
			return
		case <-ctx.Done():
			return
		case <-retryTick.C:
		}
	}
}
func (session *streamSession) Shutdown() {
	close(session.cancel)
	<-session.done
}
func (session *streamSession) waitForReadiness(ctx context.Context, streamID string) {
	waitForReadyTicker := time.NewTicker(10 * time.Second)
	defer waitForReadyTicker.Stop()
	for {
		_, err := session.kv.List(ctx)
		if err == nil {
			_, err := session.messages.GetStream(ctx, streamID)
			if err == nil {
				return
			} else {
				session.logger.Warn("messages service is not ready yet, waiting")
			}
		} else {
			session.logger.Warn("session service is not ready yet, waiting")
		}
		select {
		case <-session.cancel:
			return
		case <-ctx.Done():
			return
		case <-waitForReadyTicker.C:
		}
	}
}
func (session *streamSession) Start(ctx context.Context, streamID string, f ShardConsumer) {
	session.cancel = make(chan struct{})
	session.waitForReadiness(ctx, streamID)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		heartbeatTicker := time.NewTicker(2 * time.Second)
		purgeTicker := time.NewTicker(20 * time.Second)
		rebalanceTicker := time.NewTicker(5 * time.Second)
		wg := sync.WaitGroup{}
		session.register(ctx, streamID)
		defer unregisterConsumer(ctx, session.kv, streamID, session.GroupID, session.ConsumerID)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer heartbeatTicker.Stop()
			for {
				select {
				case <-session.cancel:
					return
				case <-ctx.Done():
					return
				case <-heartbeatTicker.C:
					err := heartbeatConsumer(ctx, session.kv, streamID, session.GroupID, session.ConsumerID)
					if err != nil {
						session.logger.Error("failed to heartbeat group", zap.Error(err))
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
					session.logger.Error("failed to purge dead consumers", zap.Error(err))
				}
				select {
				case <-session.cancel:
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
					session.logger.Error("failed to list other consumers", zap.Error(err))
					<-retryTicker.C
					continue
				}
				shards, err := session.getShards(ctx, streamID)
				if err != nil {
					session.logger.Error("failed to list shards", zap.Error(err))
					<-retryTicker.C
					continue
				}
				newAssignedShard := []string{}
				for _, shard := range getAssignedShard(shards, consumers, session.ConsumerID) {
					_, err := lockShard(ctx, session.kv, streamID, session.GroupID, shard, session.ConsumerID, 10*time.Second)
					if err != nil {
						session.logger.Warn("failed to lock shard", zap.Error(err))
						continue
					} else {
						newAssignedShard = append(newAssignedShard, shard)
					}
				}
				assignedShards = newAssignedShard
				select {
				case <-session.cancel:
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
			case <-session.cancel:
				wg.Wait()
				return
			case <-ctx.Done():
				wg.Wait()
				return
			case <-ticker.C:
				for _, shard := range assignedShards {
					err := session.consumeShard(ctx, shard)
					if err != nil {
						session.logger.Error("failed to consume shard", zap.Error(err))
					}
				}
			}
		}
	}()
}

func (session *streamSession) getShards(ctx context.Context, streamID string) ([]string, error) {
	streamConfig, err := session.messages.GetStream(ctx, streamID)
	if err != nil {
		return nil, err
	}
	return streamConfig.ShardIDs, nil
}

func (session *streamSession) consumeShard(ctx context.Context, shardId string) error {
	consumptionStart := time.Now()
	logger := session.logger.WithOptions(zap.Fields(zap.String("shard_id", shardId)))
	offset, err := session.getOffset(ctx, shardId)
	if err != nil {
		logger.Error("failed to get offset for shard", zap.Error(err))
		return err
	}
	next, messages, err := session.messages.GetMessages(ctx, session.StreamID, shardId, offset, session.MaxBatchSize)
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
	session.logger.Info("shard messages consumed",
		zap.Uint64("shard_offset", offset),
		zap.Int("shard_message_count", len(messages)),
		zap.Duration("shard_iterator_age", iteratorAge),
		zap.Duration("shard_consumption_time", time.Since(consumptionStart)),
		zap.Duration("shard_consumption_processor_time", processorElapsed),
		zap.Duration("shard_consumption_integration_time", integrationElapsed),
	)
	return session.saveOffset(ctx, shardId, next)
}
