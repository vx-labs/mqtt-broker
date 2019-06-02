package broker

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/vx-labs/mqtt-broker/sessions"

	"github.com/vx-labs/mqtt-broker/broker/listener"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions"
	topics "github.com/vx-labs/mqtt-broker/topics"
	"github.com/vx-labs/mqtt-protocol/packet"
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
func (b *Broker) OnSubscribe(transportSession *listener.Session, sess sessions.Session, p *packet.Subscribe) error {
	for idx, pattern := range p.Topic {
		subID := makeSubID(sess.ID, pattern)
		event := subscriptions.Subscription{
			Metadata: subscriptions.Metadata{
				ID:        subID,
				Pattern:   pattern,
				Qos:       p.Qos[idx],
				Tenant:    sess.Tenant,
				SessionID: sess.ID,
				Peer:      b.ID,
			},
		}
		err := b.Subscriptions.Create(event, func(publish packet.Publish) error {
			return transportSession.Publish(&publish)
		})
		if err != nil {
			return err
		}
		// Look for retained messages
		set, err := b.Topics.ByTopicPattern(sess.Tenant, pattern)
		if err != nil {
			return err
		}
		packetQoS := p.Qos[idx]
		go func() {
			set.Apply(func(message topics.RetainedMessage) {
				qos := getLowerQoS(message.Qos, packetQoS)
				transportSession.Publish(&packet.Publish{
					Header: &packet.Header{
						Qos:    qos,
						Retain: true,
					},
					Topic:   message.Topic,
					Payload: message.Payload,
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
	set = set.Filter(func(sub subscriptions.Subscription) bool {
		for _, topic := range packet.Topic {
			if bytes.Compare(topic, sub.Pattern) == 0 {
				return true
			}
		}
		return false
	})
	set.Apply(func(sub subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})
	return nil
}

func (b *Broker) deleteSessionSubscriptions(sess sessions.Session) error {
	set, err := b.Subscriptions.BySession(sess.ID)
	if err != nil {
		return err
	}
	set.Apply(func(sub subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})
	return nil
}
func (b *Broker) OnSessionClosed(sess sessions.Session) {
	err := b.deleteSessionSubscriptions(sess)
	if err != nil {
		log.Printf("WARN: failed to delete session subscriptions: %v", err)
	}
	b.Sessions.Delete(sess.ID, "session_closed")
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
	clientId := string(connectPkt.ClientId)
	tenant := transportSession.Tenant()
	transport := transportSession.TransportName()
	log.Printf("DEBUG: session %s: checking if session client-id is free", id)
	set, err := b.Sessions.ByClientID(clientId)
	if err != nil {
		return packet.CONNACK_REFUSED_SERVER_UNAVAILABLE, err
	}
	if len(set) == 0 {
		log.Printf("DEBUG: session %s: session client-id is free", id)
	} else {
		log.Printf("DEBUG: session %s: session client-id is not free, closing old sessions", id)
		if err := set.ApplyE(func(session sessions.Session) error {
			if session.Transport != nil && session.Peer == b.ID {
				log.Printf("INFO: closing old session %s", session.ID)
				return session.Transport.Close()
			}
			return nil
		}); err != nil {
			return packet.CONNACK_REFUSED_IDENTIFIER_REJECTED, err
		}
	}
	sess := sessions.Session{
		Metadata: sessions.Metadata{
			ID:                id,
			ClientID:          clientId,
			Created:           time.Now().Unix(),
			Tenant:            tenant,
			Peer:              b.ID,
			WillPayload:       connectPkt.WillPayload,
			WillQoS:           connectPkt.WillQos,
			WillRetain:        connectPkt.WillRetain,
			WillTopic:         connectPkt.WillTopic,
			Transport:         transport,
			RemoteAddress:     transportSession.RemoteAddress(),
			KeepaliveInterval: connectPkt.KeepaliveTimer,
		},
	}
	log.Printf("DEBUG: session %s: subscribing to session events", id)
	var cancels []func()
	cancels = []func(){
		transportSession.OnSubscribe(func(p *packet.Subscribe) {
			b.workers.Call(func() error {
				err := b.OnSubscribe(transportSession, sess, p)
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
				return nil
			})
		}),
		transportSession.OnPublish(func(p *packet.Publish) {
			b.workers.Call(func() error {
				err := b.OnPublish(sess, p)
				if p.Header.Qos == 0 {
					return nil
				}
				if err != nil {
					log.Printf("ERR: failed to handle message publish: %v", err)
					return err
				}
				return nil
			})
		}),
		transportSession.OnClosed(func() {
			b.workers.Call(func() error {
				b.Sessions.Delete(sess.ID, "session_disconnected")
				err := b.deleteSessionSubscriptions(sess)
				if err != nil {
					log.Printf("WARN: failed to delete session subscriptions: %v", err)
					return err
				}
				for _, cancel := range cancels {
					cancel()
				}
				return nil
			})
		}),
		transportSession.OnLost(func() {
			b.workers.Call(func() error {
				b.Sessions.Delete(sess.ID, "session_lost")
				err := b.deleteSessionSubscriptions(sess)
				if err != nil {
					log.Printf("WARN: failed to delete session subscriptions: %v", err)
				}
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
				for _, cancel := range cancels {
					cancel()
				}
				return nil
			})
		}),
	}
	log.Printf("DEBUG: session %s: creating session in distributed store", id)
	err = b.Sessions.Upsert(sess, transportSession)
	if err != nil {
		return packet.CONNACK_REFUSED_SERVER_UNAVAILABLE, err
	}
	log.Printf("INFO: session %s started", sess.ID)
	return packet.CONNACK_CONNECTION_ACCEPTED, nil
}
func (b *Broker) OnPublish(sess sessions.Session, p *packet.Publish) error {
	if b.STANOutput != nil {
		b.STANOutput <- STANMessage{
			Timestamp: time.Now(),
			Tenant:    sess.Tenant,
			Payload:   p.Payload,
			Topic:     p.Topic,
		}
	}
	if p.Header.Retain {
		message := topics.RetainedMessage{
			Metadata: topics.Metadata{
				Payload: p.Payload,
				Qos:     p.Header.Qos,
				Tenant:  sess.Tenant,
				Topic:   p.Topic,
			},
		}
		err := b.Topics.Create(message)
		if err != nil {
			log.Printf("WARN: failed to save retained message: %v", err)
		}
	}
	recipients, err := b.Subscriptions.ByTopic(sess.Tenant, p.Topic)
	if err != nil {
		return err
	}

	message := *p
	recipients.Apply(func(sub subscriptions.Subscription) {
		sub.Sender(message)
	})
	return nil
}

func (b *Broker) Authenticate(transport listener.Transport, sessionID []byte, username string, password string) (tenant string, id string, err error) {
	tenant, err = b.authHelper(transport, sessionID, username, password)
	if err != nil {
		log.Printf("WARN: authentication failed from %s: %v", transport.RemoteAddress(), err)
	}
	return tenant, uuid.New().String(), err
}
