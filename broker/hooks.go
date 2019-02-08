package broker

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/vx-labs/mqtt-broker/sessions"

	proto "github.com/golang/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/broker/listener"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions"
	topics "github.com/vx-labs/mqtt-broker/topics"
	"github.com/vx-labs/mqtt-protocol/packet"
	"github.com/weaveworks/mesh"
)

func makeSubID(session string, pattern []byte) string {
	hash := sha1.New()
	_, err := hash.Write([]byte(session))
	if err != nil {
		return ""
	}
	_, err = hash.Write(pattern)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getLowerQoS(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func (b *Broker) OnSubscribe(sess sessions.Session, packet *packet.Subscribe) error {
	for idx, pattern := range packet.Topic {
		subID := makeSubID(sess.ID, pattern)
		event := &subscriptions.Subscription{
			ID:        subID,
			Pattern:   pattern,
			Qos:       packet.Qos[idx],
			Tenant:    sess.Tenant,
			SessionID: sess.ID,
			Peer:      uint64(b.Peer.Name()),
		}
		err := b.Subscriptions.Create(event)
		if err != nil {
			return err
		}
		// Look for retained messages
		set, err := b.Topics.ByTopicPattern(sess.Tenant, pattern)
		if err != nil {
			return err
		}
		packetQoS := packet.Qos[idx]
		go func() {
			set.Apply(func(message *topics.RetainedMessage) {
				qos := getLowerQoS(message.Qos, packetQoS)
				b.dispatch(&MessagePublished{
					Payload:   message.Payload,
					Retained:  true,
					Recipient: []string{sess.ID},
					Topic:     message.Topic,
					Qos:       []int32{qos},
				})
			})
		}()
	}
	return nil
}
func (b *Broker) OnUnsubscribe(sess sessions.Session, packet *packet.Unsubscribe) error {
	set, err := b.Subscriptions.BySession(sess.ID)
	if err != nil {
		return err
	}
	set = set.Filter(func(sub *subscriptions.Subscription) bool {
		for _, topic := range packet.Topic {
			if bytes.Compare(topic, sub.Pattern) == 0 {
				return true
			}
		}
		return false
	})
	set.Apply(func(sub *subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})
	return nil
}

func (b *Broker) deleteSessionSubscriptions(sess sessions.Session) error {
	set, err := b.Subscriptions.BySession(sess.ID)
	if err != nil {
		return err
	}
	set.Apply(func(sub *subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})
	return nil
}
func (b *Broker) OnSessionClosed(sess sessions.Session) {
	err := b.deleteSessionSubscriptions(sess)
	if err != nil {
		log.Printf("WARN: failed to delete session subscriptions: %v", err)
	}
	b.Sessions.Delete(sess.ID, "session_disconnected")
	return
}
func (b *Broker) OnSessionLost(sess sessions.Session) {
	if len(sess.WillTopic) > 0 {
		b.OnPublish(sess, &packet.Publish{
			Header: &packet.Header{
				Dup:    false,
				Retain: sess.WillRetain,
				Qos:    sess.WillQoS,
			},
			Payload: sess.WillPayload,
			Topic:   sess.WillTopic,
		})
	}
	err := b.deleteSessionSubscriptions(sess)
	if err != nil {
		log.Printf("WARN: failed to delete session subscriptions: %v", err)
	}
	b.Sessions.Delete(sess.ID, "session_lost")
}

func (b *Broker) OnConnect(transportSession *listener.Session) (int32, error) {
	connectPkt := transportSession.Connect()
	id := transportSession.ID()
	tenant := transportSession.Tenant()
	transport := transportSession.TransportName()
	_, err := b.Sessions.ByID(id)
	if err == nil {
		return packet.CONNACK_REFUSED_IDENTIFIER_REJECTED, errors.New("sessions already exist")
	}
	sess := sessions.Session{
		ID:                id,
		ClientID:          connectPkt.ClientId,
		Created:           time.Now().Unix(),
		Tenant:            tenant,
		Peer:              uint64(b.Peer.Name()),
		WillPayload:       connectPkt.WillPayload,
		WillQoS:           connectPkt.WillQos,
		WillRetain:        connectPkt.WillRetain,
		WillTopic:         connectPkt.WillTopic,
		Transport:         transport,
		RemoteAddress:     transportSession.RemoteAddress(),
		KeepaliveInterval: connectPkt.KeepaliveTimer,
	}
	err = b.Sessions.Upsert(&sess)
	if err != nil {
		return packet.CONNACK_REFUSED_SERVER_UNAVAILABLE, err
	}
	var cancels []func()
	cancels = []func(){
		transportSession.OnSubscribe(func(p *packet.Subscribe) {
			err := b.OnSubscribe(sess, p)
			if err == nil {
				qos := make([]int32, len(p.Qos))

				// QoS2 is not supported for now
				for idx := range p.Qos {
					if p.Qos[idx] > 1 {
						qos[idx] = 1
					} else {
						qos[idx] = p.Qos[idx]
					}
				}
				transportSession.SubAck(p.MessageId, qos)
			}
		}),
		transportSession.OnPublish(func(p *packet.Publish) {
			err := b.OnPublish(sess, p)
			if p.Header.Qos == 0 {
				return
			}
			if err != nil {
				log.Printf("ERR: failed to handle message publish: %v", err)
				return
			}
		}),
		transportSession.OnClosed(func() {
			b.Sessions.Delete(sess.ID, "session_disconnected")
		}),
		transportSession.OnLost(func() {
			b.Sessions.Delete(sess.ID, "session_lost")
		}),
		b.OnMessagePublished(transportSession.ID(), func(p *packet.Publish) {
			transportSession.Publish(p)
		}),
		b.Sessions.On(sessions.SessionCreated+"/"+id, func(s *sessions.Session) {
			transportSession.Close()
		}),
		b.Sessions.On(sessions.SessionDeleted+"/"+id, func(s *sessions.Session) {
			err := b.deleteSessionSubscriptions(sess)
			if err != nil {
				log.Printf("WARN: failed to delete session subscriptions: %v", err)
			}
			for _, cancel := range cancels {
				cancel()
			}
			transportSession.Close()
			if s.ClosureReason == "session_lost" {
				if len(sess.WillTopic) > 0 {
					b.OnPublish(sess, &packet.Publish{
						Header: &packet.Header{
							Dup:    false,
							Retain: sess.WillRetain,
							Qos:    sess.WillQoS,
						},
						Payload: sess.WillPayload,
						Topic:   sess.WillTopic,
					})
				}
			}
		}),
	}
	return packet.CONNACK_CONNECTION_ACCEPTED, nil
}
func (b *Broker) OnPublish(sess sessions.Session, packet *packet.Publish) error {
	if b.STANOutput != nil {
		b.STANOutput <- STANMessage{
			Timestamp: time.Now(),
			Tenant:    sess.Tenant,
			Payload:   packet.Payload,
			Topic:     packet.Topic,
		}
	}
	if packet.Header.Retain {
		message := &topics.RetainedMessage{
			Payload: packet.Payload,
			Qos:     packet.Header.Qos,
			Tenant:  sess.Tenant,
			Topic:   packet.Topic,
		}
		err := b.Topics.Create(message)
		if err != nil {
			log.Printf("WARN: failed to save retained message: %v", err)
		}
	}
	recipients, err := b.Subscriptions.ByTopic(sess.Tenant, packet.Topic)
	if err != nil {
		return err
	}

	peers := map[uint64]*MessagePublished{}
	recipients.Apply(func(sub *subscriptions.Subscription) {
		if _, ok := peers[sub.Peer]; !ok {
			peers[sub.Peer] = &MessagePublished{
				Payload:   packet.Payload,
				Qos:       make([]int32, 0, len(recipients.Subscriptions)),
				Recipient: make([]string, 0, len(recipients.Subscriptions)),
				Topic:     packet.Topic,
			}
		}
		peers[sub.Peer].Recipient = append(peers[sub.Peer].Recipient, sub.SessionID)
		qos := packet.Header.Qos
		if qos > sub.Qos {
			qos = sub.Qos
		}
		peers[sub.Peer].Qos = append(peers[sub.Peer].Qos, qos)
	})
	for peer, message := range peers {
		payload, err := proto.Marshal(message)
		if err != nil {
			return err
		}
		if peer == uint64(b.Peer.Name()) {
			b.dispatch(message)
		} else {
			b.Peer.Send(mesh.PeerName(peer), payload)
		}
	}
	return nil
}

func (b *Broker) Authenticate(transport listener.Transport, sessionID []byte, username string, password string) (tenant string, id string, err error) {
	tenant, id, err = b.authHelper(transport, sessionID, username, password)
	if err != nil {
		log.Printf("WARN: authentication failed from %s: %v", transport.RemoteAddress(), err)
	}
	return tenant, id, err
}
