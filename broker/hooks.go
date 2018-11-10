package broker

import (
	"bytes"
	"crypto/sha1"
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

func (b *Broker) OnSubscribe(id string, tenant string, packet *packet.Subscribe) error {
	for idx, pattern := range packet.Topic {
		subID := makeSubID(id, pattern)
		event := &subscriptions.Subscription{
			ID:        subID,
			Pattern:   pattern,
			Qos:       packet.Qos[idx],
			Tenant:    tenant,
			SessionID: id,
			Peer:      uint64(b.Peer.Name()),
		}
		err := b.Subscriptions.Create(event)
		if err != nil {
			return err
		}
		// Look for retained messages
		set, err := b.Topics.ByTopicPattern(tenant, pattern)
		if err != nil {
			return err
		}
		go func() {
			set.Apply(func(message *topics.RetainedMessage) {
				qos := message.Qos
				if packet.Qos[idx] < qos {
					qos = packet.Qos[idx]
				}
				b.dispatch(&MessagePublished{
					Payload:   message.Payload,
					Retained:  true,
					Recipient: []string{id},
					Topic:     message.Topic,
					Qos:       []int32{qos},
				})
			})
		}()
	}
	return nil
}
func (b *Broker) OnUnsubscribe(id string, tenant string, packet *packet.Unsubscribe) error {
	set, err := b.Subscriptions.BySession(id)
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

func (b *Broker) OnSessionClosed(id, tenant string) {
	set, err := b.Subscriptions.BySession(id)
	if err != nil {
		return
	}
	sess, err := b.Sessions.ByID(id)
	if err != nil || sess.Peer != uint64(b.Peer.Name()) {
		return
	}
	set.Apply(func(sub *subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})
	b.Sessions.Delete(sess.ID)
	return
}
func (b *Broker) OnSessionLost(id, tenant string) {
	defer b.OnSessionClosed(id, tenant)
	sess, err := b.Sessions.ByID(id)
	if err != nil {
		return
	}
	if len(sess.WillTopic) > 0 {
		b.OnPublish(id, tenant, &packet.Publish{
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

func (b *Broker) OnConnect(transportSession *listener.Session) {
	connectPkt := transportSession.Connect()
	id := transportSession.ID()
	tenant := transportSession.Tenant()
	transport := transportSession.TransportName()
	sess := &sessions.Session{
		ID:          id,
		ClientID:    connectPkt.ClientId,
		Created:     time.Now().Unix(),
		Tenant:      tenant,
		Peer:        uint64(b.Peer.Name()),
		WillPayload: connectPkt.WillPayload,
		WillQoS:     connectPkt.WillQos,
		WillRetain:  connectPkt.WillRetain,
		WillTopic:   connectPkt.WillTopic,
		Transport:   transport,
	}
	b.Sessions.Upsert(sess)
	var cancels []func()
	cancels = []func(){
		transportSession.OnSubscribe(func(p *packet.Subscribe) {
			err := b.OnSubscribe(id, tenant, p)
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
			err := b.OnPublish(transportSession.ID(), transportSession.Tenant(), p)
			if p.Header.Qos == 0 {
				return
			}
			if err != nil {
				log.Printf("ERR: failed to handle message publish: %v", err)
				return
			}
			if p.Header.Qos == 1 {
				transportSession.PubAck(p.MessageId)
			}
		}),
		transportSession.OnClosed(func() {
			b.OnSessionClosed(transportSession.ID(), transportSession.Tenant())
			for _, cancel := range cancels {
				cancel()
			}
		}),
		transportSession.OnLost(func() {
			b.OnSessionLost(transportSession.ID(), transportSession.Tenant())
			for _, cancel := range cancels {
				cancel()
			}
		}),
		b.OnMessagePublished(transportSession.ID(), func(p *packet.Publish) {
			transportSession.Publish(p)
		}),
	}
}
func (b *Broker) OnPublish(id, tenant string, packet *packet.Publish) error {
	if packet.Header.Retain {
		message := &topics.RetainedMessage{
			Payload: packet.Payload,
			Qos:     packet.Header.Qos,
			Tenant:  tenant,
			Topic:   packet.Topic,
		}
		err := b.Topics.Create(message)
		if err != nil {
			log.Printf("WARN: failed to save retained message: %v", err)
		}
	}
	recipients, err := b.Subscriptions.ByTopic(tenant, packet.Topic)
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

func (b *Broker) Authenticate(transport listener.Transport, sessionID, username string, password string) (tenant string, id string, err error) {
	tenant, id, err = b.authHelper(transport, sessionID, username, password)
	if err != nil {
		log.Printf("WARN: authentication failed from %s: %v", transport.RemoteAddress(), err)
	}
	return tenant, id, err
}
