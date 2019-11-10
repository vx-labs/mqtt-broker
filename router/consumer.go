package router

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/pkg/errors"
	"github.com/vx-labs/mqtt-broker/broker/pb"
	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	queues "github.com/vx-labs/mqtt-broker/queues/pb"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions/pb"
	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/gogo/protobuf/proto"
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
	if next == offset {
		return
	}
	b.logger.Debug("consuming shard messages", zap.String("shard_id", shardId), zap.Uint64("shard_offset", offset))
	idx, err := b.v2ConsumePayload(messages)
	if err != nil {
		idx, err = b.v1ConsumePayload(messages[idx:])
		if err != nil {
			b.saveOffset(shardId, messages[idx].Offset)
			b.logger.Error("failed to consume shard", zap.String("shard_id", shardId), zap.Uint64("shard_offset", messages[idx].Offset), zap.Error(err))
			return
		} else {
			b.logger.Debug("shard messages consumed", zap.String("formation_version", "v1"), zap.String("shard_id", shardId), zap.Uint64("shard_offset", offset))
		}
	} else {
		b.logger.Debug("shard messages consumed", zap.String("formation_version", "v2"), zap.String("shard_id", shardId), zap.Uint64("shard_offset", offset))
	}
	b.saveOffset(shardId, next)
}

func (b *server) v1ConsumePayload(messages []*messages.StoredMessage) (int, error) {
	publishes := make([]*packet.Publish, len(messages))
	topics := RouteSet{}
	for idx := range messages {
		publishes[idx] = &packet.Publish{}
		err := proto.Unmarshal(messages[idx].Payload, publishes[idx])
		if err != nil {
			return idx, errors.Wrap(err, "failed to decode message for shard")
		}
	}

	for idx, message := range publishes {
		if !contains(message.Topic, topics) {
			recipients, err := b.Subscriptions.ByTopic(b.ctx, "_default", message.Topic)
			if err != nil {
				return idx, errors.Wrap(err, "failed to resolve recipients for message")
			}
			topics = append(topics, Route{
				Topic:      message.Topic,
				Recipients: recipients,
			})
		}
	}
	for idx := range messages {
		p := publishes[idx]
		p.Header.Retain = false
		recipients := topics.Recipients(p.Topic)

		payload := make([]queues.MessageBatch, len(recipients))
		for idx := range recipients {
			payload[idx] = queues.MessageBatch{ID: recipients[idx].SessionID, Publish: p}
		}
		err := b.Queues.PutMessageBatch(b.ctx, payload)
		if err != nil {
			return idx, errors.Wrap(err, "failed to enqueue message")
		}
	}
	return 0, nil
}

func (b *server) v2ConsumePayload(messages []*messages.StoredMessage) (int, error) {
	publishes := make([]*pb.MessagePublished, len(messages))
	topics := RouteSet{}
	for idx := range messages {
		publishes[idx] = &pb.MessagePublished{}
		err := proto.Unmarshal(messages[idx].Payload, publishes[idx])
		if err != nil {
			return idx, errors.Wrap(err, "failed to decode message for shard")
		}
	}

	for idx, message := range publishes {
		if message.Publish == nil {
			return idx, errors.New("received invalid encoded payload")
		}
		if !contains(message.Publish.Topic, topics) {
			recipients, err := b.Subscriptions.ByTopic(b.ctx, message.Tenant, message.Publish.Topic)
			if err != nil {
				return idx, errors.Wrap(err, "failed to resolve recipients for message")
			}
			topics = append(topics, Route{
				Topic:      message.Publish.Topic,
				Recipients: recipients,
			})
		}
	}
	for idx := range messages {
		p := publishes[idx].Publish
		p.Header.Retain = false
		recipients := topics.Recipients(p.Topic)

		payload := make([]queues.MessageBatch, len(recipients))
		for idx := range recipients {
			payload[idx] = queues.MessageBatch{ID: recipients[idx].SessionID, Publish: p}
		}
		err := b.Queues.PutMessageBatch(b.ctx, payload)
		if err != nil {
			return idx, errors.Wrap(err, "failed to enqueue message")
		}
	}
	return 0, nil
}
