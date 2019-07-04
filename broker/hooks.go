package broker

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/transport"

	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions"
	topics "github.com/vx-labs/mqtt-broker/topics"
	"github.com/vx-labs/mqtt-protocol/packet"
)

func newUUID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to read random bytes: %v", err))
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

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
	message.Header.Retain = false
	recipients.Apply(func(sub subscriptions.Subscription) {
		sub.Sender(b.ctx, message)
	})
	return nil
}

func (b *Broker) Authenticate(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
	tenant, err = b.authHelper(transport, sessionID, username, password)
	if err != nil {
		log.Printf("WARN: authentication failed from %s: %v", transport.RemoteAddress, err)
	}
	return tenant, err
}
