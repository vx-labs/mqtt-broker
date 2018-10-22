package broker

import (
	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
	"github.com/vx-labs/mqtt-broker/topics"
)

func (b *Broker) ListSessions() (sessions.SessionList, error) {
	return b.Sessions.All()
}
func (b *Broker) ListSubscriptions() (subscriptions.SubscriptionList, error) {
	return b.Subscriptions.All()
}
func (b *Broker) ListRetainedMessages() (topics.RetainedMessageList, error) {
	return b.Topics.All()
}
