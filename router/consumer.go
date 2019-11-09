package router

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	queues "github.com/vx-labs/mqtt-broker/queues/pb"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions/pb"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

func (b *server) getShards(streamId string) ([]string, error) {
	streamConfig, err := b.Messages.GetStream(b.ctx, streamId)
	if err != nil {
		return nil, err
	}
	return streamConfig.ShardIDs, nil
}

func (b *server) saveOffset(shard string, offset uint64) error {
	return b.KV.Set(b.ctx, []byte(fmt.Sprintf("router/offset/%s", shard)), uint64ToBytes(offset))
}
func (b *server) getOffset(shard string) (uint64, error) {
	out, err := b.KV.Get(b.ctx, []byte(fmt.Sprintf("router/offset/%s", shard)))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return 0, nil
			}
		}
		return 0, err
	}
	return binary.BigEndian.Uint64(out), nil
}
func (b *server) consumeMessages() {
	b.logger.Debug("started message consumer")
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		<-ticker.C
		shards, err := b.getShards("messages")
		if err != nil {
			continue
		}
		for _, shard := range shards {
			b.consumeShard("messages", shard)
		}
	}
}

type RouteSet []Route

func (r RouteSet) Recipients(topic []byte) []*subscriptions.Metadata {
	for _, route := range r {
		if bytes.Compare(route.Topic, topic) == 0 {
			return route.Recipients
		}
	}
	return nil
}

type Route struct {
	Topic      []byte
	Recipients []*subscriptions.Metadata
}

func contains(needle []byte, slice []Route) bool {
	for idx := range slice {
		if bytes.Compare(needle, slice[idx].Topic) == 0 {
			return true
		}
	}
	return false
}

func (b *server) consumeShard(streamId string, shardId string) {
	offset, err := b.getOffset(shardId)
	if err != nil {
		b.logger.Error("failed to get offset for shard", zap.String("shard_id", shardId), zap.Error(err))
		return
	}
	next, messages, err := b.Messages.GetMessages(b.ctx, streamId, shardId, offset, 100)
	if err != nil {
		b.logger.Error("failed to get messages for shard", zap.String("shard_id", shardId), zap.Error(err))
		return
	}
	processedOffset := offset
	if next == offset {
		return
	}
	b.logger.Debug("consuming shard messages", zap.String("shard_id", shardId), zap.Uint64("shard_offset", offset))
	publishes := make([]*packet.Publish, len(messages))
	topics := RouteSet{}
	for idx := range messages {
		publishes[idx] = &packet.Publish{}
		err := proto.Unmarshal(messages[idx].Payload, publishes[idx])
		if err != nil {
			b.saveOffset(shardId, processedOffset)
			b.logger.Error("failed to decode message for shard", zap.String("shard_id", shardId), zap.Error(err))
			return
		}
	}

	for _, message := range publishes {
		if !contains(message.Topic, topics) {
			recipients, err := b.Subscriptions.ByTopic(b.ctx, "_default", message.Topic)
			if err != nil {
				b.saveOffset(shardId, processedOffset)
				b.logger.Error("failed to resolve recipients for message", zap.String("shard_id", shardId), zap.Error(err))
			}
			topics = append(topics, Route{
				Topic:      message.Topic,
				Recipients: recipients,
			})
		}
	}
	for idx, message := range messages {
		p := publishes[idx]
		recipients := topics.Recipients(p.Topic)

		payload := make([]queues.MessageBatch, len(recipients))
		for idx := range recipients {
			payload[idx] = queues.MessageBatch{ID: recipients[idx].SessionID, Publish: p}
		}
		err = b.Queues.PutMessageBatch(b.ctx, payload)
		if err != nil {
			b.logger.Error("failed to enqueue message", zap.String("shard_id", shardId), zap.Error(err))
		}
		b.saveOffset(shardId, message.Offset)
	}
	b.saveOffset(shardId, next)
	b.logger.Debug("shard messages consumed", zap.String("shard_id", shardId), zap.Uint64("shard_offset", offset))
}
