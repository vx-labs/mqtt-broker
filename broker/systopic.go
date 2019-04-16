package broker

import (
	"encoding/json"
	"fmt"

	"github.com/vx-labs/mqtt-broker/broker/rpc"
	"github.com/vx-labs/mqtt-broker/topics"

	"github.com/vx-labs/mqtt-broker/peers"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
)

func (b *Broker) retainMessage(tenant string, topic []byte, payload []byte, qos int32) error {
	return b.Topics.Create(&topics.RetainedMessage{
		Tenant:  tenant,
		Topic:   topic,
		Payload: payload,
		Qos:     qos,
	})
}
func (b *Broker) dispatchToLocalSessions(tenant string, topic []byte, payload []byte, defaultQoS int32) {
	recipients, qos := b.resolveRecipients(tenant, topic, defaultQoS)
	message := &rpc.MessagePublished{
		Payload:   payload,
		Recipient: recipients,
		Qos:       qos,
		Topic:     topic,
	}
	b.dispatch(message)
}

func (b *Broker) RetainThenDispatchToLocalSessions(tenant string, topic []byte, payload []byte, qos int32) {
	if b.retainMessage("_default", topic, payload, 1) == nil {
		b.dispatchToLocalSessions("_default", topic, payload, 1)
	}
}
func (b *Broker) setupSYSTopic() {
	b.Peers.On(peers.PeerCreated, func(s *peers.Peer) {
		payload, err := json.Marshal(s)
		if err == nil {
			topic := []byte(fmt.Sprintf("$SYS/peers/%s", s.ID))
			b.RetainThenDispatchToLocalSessions("_default", topic, payload, 1)
		}
	})
	b.Peers.On(peers.PeerDeleted, func(s *peers.Peer) {
		topic := []byte(fmt.Sprintf("$SYS/peers/%s", s.ID))
		b.RetainThenDispatchToLocalSessions("_default", topic, nil, 1)
	})
	b.Subscriptions.On(subscriptions.SubscriptionCreated, func(s *subscriptions.Subscription) {
		payload, err := json.Marshal(s)
		if err == nil {
			topic := []byte(fmt.Sprintf("$SYS/subscriptions/%s", s.ID))
			b.RetainThenDispatchToLocalSessions(s.Tenant, topic, payload, 1)
		}
	})
	b.Subscriptions.On(subscriptions.SubscriptionDeleted, func(s *subscriptions.Subscription) {
		topic := []byte(fmt.Sprintf("$SYS/subscriptions/%s", s.ID))
		b.RetainThenDispatchToLocalSessions(s.Tenant, topic, nil, 1)
	})

	b.Sessions.On(sessions.SessionCreated, func(s sessions.Session) {
		payload, err := json.Marshal(s)
		if err == nil {
			topic := []byte(fmt.Sprintf("$SYS/sessions/%s", s.ID))
			b.RetainThenDispatchToLocalSessions(s.Tenant, topic, payload, 1)
		}
	})
	b.Sessions.On(sessions.SessionDeleted, func(s sessions.Session) {
		topic := []byte(fmt.Sprintf("$SYS/sessions/%s", s.ID))
		b.RetainThenDispatchToLocalSessions(s.Tenant, topic, nil, 1)
	})
}
