package router

import (
	"bytes"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/vx-labs/mqtt-broker/services/broker/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	queues "github.com/vx-labs/mqtt-broker/services/queues/pb"
	subscriptions "github.com/vx-labs/mqtt-broker/services/subscriptions/pb"
	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/gogo/protobuf/proto"
)

type RouteSet []Route

func (r RouteSet) Recipients(topic []byte) []*subscriptions.Subscription {
	for _, route := range r {
		if bytes.Compare(route.Topic, topic) == 0 {
			return route.Recipients
		}
	}
	return nil
}

type Route struct {
	Topic      []byte
	Recipients []*subscriptions.Subscription
}

func contains(needle []byte, slice []Route) bool {
	for idx := range slice {
		if bytes.Compare(needle, slice[idx].Topic) == 0 {
			return true
		}
	}
	return false
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
	count := 0
	for idx := range messages {
		p := publishes[idx].Publish
		count += len(topics.Recipients(p.Topic))
	}
	payload := make([]queues.MessageBatch, count)
	offset := 0
	for idx := range messages {
		p := publishes[idx].Publish
		p.Header.Retain = false
		recipients := topics.Recipients(p.Topic)
		for idx := range recipients {
			m := &packet.Publish{
				Payload: p.Payload,
				Topic:   p.Topic,
				Header: &packet.Header{
					Dup: p.Header.Dup,
					Qos: p.Header.Qos,
				},
			}
			if m.Header.Qos > recipients[idx].Qos {
				m.Header.Qos = recipients[idx].Qos
			}
			payload[offset] = queues.MessageBatch{ID: recipients[idx].SessionID, Publish: m}
			offset++
		}
	}
	if len(payload) > 0 {
		err := b.Queues.PutMessageBatch(b.ctx, payload)
		if err != nil {
			b.logger.Error("failed to enqueue messages", zap.Error(err))
		}
		b.logger.Info("enqueued messages", zap.Int("message_count", count))
	}
	return 0, nil
}
