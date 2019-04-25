package broker

import (
	"errors"

	"github.com/vx-labs/mqtt-broker/broker/rpc"
	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
	"github.com/vx-labs/mqtt-broker/topics"
)

func (b *Broker) ListSessions() (sessions.SessionSet, error) {
	return b.Sessions.All()
}
func (b *Broker) ListSubscriptions() (subscriptions.SubscriptionSet, error) {
	return b.Subscriptions.All()
}
func (b *Broker) ListRetainedMessages() (topics.RetainedMessageSet, error) {
	return b.Topics.All()
}
func (b *Broker) CloseSession(id string) error {
	session, err := b.Sessions.ByID(id)
	if err != nil {
		return err
	}
	if session.Transport != nil {
		return session.Transport.Close()
	}
	return errors.New("session not managed on this node")
}
func (b *Broker) DistributeMessage(message *rpc.MessagePublished) error {
	b.dispatch(message)
	return nil
}
