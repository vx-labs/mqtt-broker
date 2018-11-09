package broker

import (
	"fmt"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
	"github.com/vx-labs/mqtt-protocol/packet"
)

func (b *Broker) setupSYSTopic() {
	b.Subscriptions.On(subscriptions.SubscriptionCreated, func(s *subscriptions.Subscription) {
		if s.Peer == uint64(b.Peer.Name()) {
			b.OnPublish("_sys", s.Tenant, &packet.Publish{
				Header:    &packet.Header{},
				MessageId: 1,
				Payload:   []byte(fmt.Sprintf("%s subscribed to %s", s.SessionID, string(s.Pattern))),
				Topic:     []byte("$SYS/events/session_subscribed"),
			})
		}
	})
	b.Subscriptions.On(subscriptions.SubscriptionDeleted, func(s *subscriptions.Subscription) {
		if s.Peer == uint64(b.Peer.Name()) {
			b.OnPublish("_sys", s.Tenant, &packet.Publish{
				Header:    &packet.Header{},
				MessageId: 1,
				Payload:   []byte(fmt.Sprintf("%s unsubscribed to %s", s.SessionID, string(s.Pattern))),
				Topic:     []byte("$SYS/events/session_unsubscribe"),
			})
		}
	})

	b.Sessions.On(sessions.SessionCreated, func(s *sessions.Session) {
		if s.Peer == uint64(b.Peer.Name()) {
			b.OnPublish("_sys", s.Tenant, &packet.Publish{
				Header:    &packet.Header{},
				MessageId: 1,
				Payload:   []byte(fmt.Sprintf("%s connected with client_id=%s", s.ID, string(s.ClientID))),
				Topic:     []byte("$SYS/events/session_created"),
			})
		}
	})
	b.Sessions.On(sessions.SessionDeleted, func(s *sessions.Session) {
		if s.Peer == uint64(b.Peer.Name()) {
			b.OnPublish("_sys", s.Tenant, &packet.Publish{
				Header:    &packet.Header{},
				MessageId: 1,
				Payload:   []byte(fmt.Sprintf("%s with client_id=%s disconnected", s.ID, string(s.ClientID))),
				Topic:     []byte("$SYS/events/session_deleted"),
			})
		}
	})
}
