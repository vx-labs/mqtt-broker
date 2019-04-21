package broker

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/vx-labs/mqtt-broker/sessions"

	"github.com/vx-labs/mqtt-broker/broker/listener"
	"github.com/vx-labs/mqtt-broker/broker/rpc"
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
func (b *Broker) OnSubscribe(sess sessions.Session, packet *packet.Subscribe) error {
	for idx, pattern := range packet.Topic {
		subID := makeSubID(sess.ID, pattern)
		event := &subscriptions.Subscription{
			ID:        subID,
			Pattern:   pattern,
			Qos:       packet.Qos[idx],
			Tenant:    sess.Tenant,
			SessionID: sess.ID,
			Peer:      b.ID,
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
				b.dispatch(&rpc.MessagePublished{
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
	set, err := b.Sessions.ByClientID(clientId)
	if err != nil {
		return packet.CONNACK_REFUSED_SERVER_UNAVAILABLE, err
	}
	if err := set.ApplyE(func(session sessions.Session) error {
		return session.Close()
	}); err != nil {
		return packet.CONNACK_REFUSED_IDENTIFIER_REJECTED, err
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
	err = b.Sessions.Upsert(sess, transportSession.Close)
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
		b.Sessions.On(sessions.SessionDeleted+"/"+id, func(s sessions.Session) {
			for _, cancel := range cancels {
				cancel()
			}
			transportSession.Close()
			err := b.deleteSessionSubscriptions(sess)
			if err != nil {
				log.Printf("WARN: failed to delete session subscriptions: %v", err)
			}
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

	peers := map[string]*rpc.MessagePublished{}
	recipients.Apply(func(sub *subscriptions.Subscription) {
		if _, ok := peers[sub.Peer]; !ok {
			peers[sub.Peer] = &rpc.MessagePublished{
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for peer, message := range peers {
		if peer == b.ID {
			b.dispatch(message)
		} else {
			addr, err := b.mesh.MemberRPCAddress(peer)
			if err != nil {
				log.Printf("WARN: unknown node found in message recipients: %s", peer)
			} else {
				err = b.RPCCaller.Call(addr, func(c rpc.BrokerServiceClient) error {
					_, err := c.DistributeMessage(ctx, message)
					return err
				})
				if err != nil {
					log.Printf("WARN: failed to send message: %v", err)
				}
			}
		}
	}
	return nil
}

func (b *Broker) Authenticate(transport listener.Transport, sessionID []byte, username string, password string) (tenant string, id string, err error) {
	tenant, err = b.authHelper(transport, sessionID, username, password)
	if err != nil {
		log.Printf("WARN: authentication failed from %s: %v", transport.RemoteAddress(), err)
	}
	return tenant, uuid.New().String(), err
}
