package broker

import (
	"encoding/json"
	"fmt"

	"github.com/vx-labs/mqtt-broker/topics"
	"github.com/vx-labs/mqtt-protocol/packet"

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
	recipients, err := b.Subscriptions.ByTopic(tenant, topic)
	if err != nil {
		return
	}
	recipients = recipients.Filter(func(sub subscriptions.Subscription) bool {
		return sub.Peer == b.mesh.ID()
	})
	message := packet.Publish{
		Payload: payload,
		Topic:   topic,
		Header: &packet.Header{
			Qos: defaultQoS,
		},
	}
	recipients.Apply(func(sub subscriptions.Subscription) {
		sub.Sender(message)
	})
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
	b.Subscriptions.On(subscriptions.SubscriptionCreated, func(s subscriptions.Subscription) {
		payload, err := json.Marshal(s)
		if err == nil {
			topic := []byte(fmt.Sprintf("$SYS/subscriptions/%s", s.ID))
			b.RetainThenDispatchToLocalSessions(s.Tenant, topic, payload, 1)
		}
	})
	b.Subscriptions.On(subscriptions.SubscriptionDeleted, func(s subscriptions.Subscription) {
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
