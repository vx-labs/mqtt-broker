package broker

import (
	"github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/subscriptions"
	"github.com/vx-labs/mqtt-broker/topics"
)

func (b *Broker) ListSubscriptions() (subscriptions.SubscriptionSet, error) {
	return b.Subscriptions.All()
}
func (b *Broker) ListRetainedMessages() (topics.RetainedMessageSet, error) {
	return b.Topics.All()
}
func (b *Broker) DistributeMessage(message *pb.MessagePublished) error {
	b.dispatch(message)
	return nil
}
