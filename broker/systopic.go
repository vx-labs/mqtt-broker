package broker

import (
	"encoding/json"
	"fmt"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
	"github.com/vx-labs/mqtt-protocol/packet"
)

func (b *Broker) setupSYSTopic() {
	b.Subscriptions.On(subscriptions.SubscriptionCreated, func(s *subscriptions.Subscription) {
		if s.Peer == uint64(b.Peer.Name()) {
			payload, err := json.Marshal(s)
			if err == nil {
				b.OnPublish("_sys", s.Tenant, &packet.Publish{
					Header: &packet.Header{
						Retain: true,
					},
					MessageId: 1,
					Payload:   payload,
					Topic:     []byte(fmt.Sprintf("$SYS/subscriptions/%s", s.ID)),
				})
			}
		}
	})
	b.Subscriptions.On(subscriptions.SubscriptionDeleted, func(s *subscriptions.Subscription) {
		if s.Peer == uint64(b.Peer.Name()) {
			b.OnPublish("_sys", s.Tenant, &packet.Publish{
				Header: &packet.Header{
					Retain: true,
				},
				MessageId: 1,
				Payload:   nil,
				Topic:     []byte(fmt.Sprintf("$SYS/subscriptions/%s", s.ID)),
			})
		}
	})

	b.Sessions.On(sessions.SessionCreated, func(s *sessions.Session) {
		if s.Peer == uint64(b.Peer.Name()) {
			payload, err := json.Marshal(s)
			if err == nil {
				b.OnPublish("_sys", s.Tenant, &packet.Publish{
					Header: &packet.Header{
						Retain: true,
					},
					MessageId: 1,
					Payload:   payload,
					Topic:     []byte(fmt.Sprintf("$SYS/sessions/%s", s.ID)),
				})
			}
		}
	})
	b.Sessions.On(sessions.SessionDeleted, func(s *sessions.Session) {
		if s.Peer == uint64(b.Peer.Name()) {
			b.OnPublish("_sys", s.Tenant, &packet.Publish{
				Header: &packet.Header{
					Retain: true,
				},
				MessageId: 1,
				Payload:   nil,
				Topic:     []byte(fmt.Sprintf("$SYS/sessions/%s", s.ID)),
			})

		}
	})
}
