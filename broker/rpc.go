package broker

import (
	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
	"github.com/vx-labs/mqtt-broker/topics"
)

func (b *Broker) ListSessions() ([]*sessions.Session, error) {
	return b.Sessions.All()
}
func (b *Broker) ListSubscriptions() ([]*subscriptions.Subscription, error) {
	return b.Subscriptions.All()
}
func (b *Broker) ListRetainedMessages() ([]*topics.RetainedMessage, error) {
	return b.Topics.All()
}
