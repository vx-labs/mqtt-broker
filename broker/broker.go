package broker

import (
	"context"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/vx-labs/mqtt-broker/transport"

	"github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/cluster"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/topics"

	"github.com/vx-labs/mqtt-protocol/packet"

	publishQueue "github.com/vx-labs/mqtt-broker/queues/publish"
	"github.com/vx-labs/mqtt-broker/subscriptions"
)

const (
	EVENT_MESSAGE_PUBLISHED = "message_published"
	EVENT_STATE_UPDATED     = "state_updated"
)

type PeerStore interface {
	ByID(id string) (cluster.Peer, error)
	All() (cluster.SubscriptionSet, error)
	Delete(id string) error
	On(event string, handler func(cluster.Peer)) func()
}
type SessionStore interface {
	ByID(id string) (sessions.Session, error)
	ByClientID(id string) (sessions.SessionSet, error)
	ByPeer(peer string) (sessions.SessionSet, error)
	All() (sessions.SessionSet, error)
	Exists(id string) bool
	Upsert(sess sessions.Session, transport sessions.Transport) error
	Delete(id, reason string) error
}

type TopicStore interface {
	Create(message topics.RetainedMessage) error
	ByTopicPattern(tenant string, pattern []byte) (topics.RetainedMessageSet, error)
	All() (topics.RetainedMessageSet, error)
}
type SubscriptionStore interface {
	ByTopic(tenant string, pattern []byte) (subscriptions.SubscriptionSet, error)
	ByID(id string) (subscriptions.Subscription, error)
	All() (subscriptions.SubscriptionSet, error)
	ByPeer(peer string) (subscriptions.SubscriptionSet, error)
	BySession(id string) (subscriptions.SubscriptionSet, error)
	Sessions() ([]string, error)
	Create(message subscriptions.Subscription, sender func(context.Context, packet.Publish) error) error
	Delete(id string) error
}
type Queue interface {
	Enqueue(p *publishQueue.Message)
	Consume(f func(*publishQueue.Message))
	Close() error
}

type Broker struct {
	ID            string
	authHelper    func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error)
	mesh          cluster.Mesh
	Subscriptions SubscriptionStore
	Sessions      SessionStore
	Topics        TopicStore
	Peers         PeerStore
	STANOutput    chan STANMessage
	workers       *Pool
	ctx           context.Context
	publishQueue  Queue
}

func (b *Broker) RemoteRPCProvider(id, peer string) sessions.Transport {
	return &localTransport{
		id:   id,
		peer: peer,
		ctx:  b.ctx,
		mesh: b.mesh,
	}
}

func New(id string, mesh cluster.Mesh, config Config) *Broker {
	ctx := context.Background()
	broker := &Broker{
		ID:           id,
		authHelper:   config.AuthHelper,
		workers:      NewPool(25),
		ctx:          ctx,
		mesh:         mesh,
		publishQueue: publishQueue.New(),
	}

	if config.NATSURL != "" {
		ch := make(chan STANMessage)
		if err := exportToSTAN(config, ch); err != nil {
			log.Printf("WARN: failed to start STAN message exporter to %s: %v", config.NATSURL, err)
		} else {
			log.Printf("INFO: started NATS message exporter")
			broker.STANOutput = ch
		}
	}
	broker.startPublishConsumers()
	return broker
}
func (broker *Broker) Start(layer cluster.ServiceLayer) {
	subscriptionsStore, err := subscriptions.NewMemDBStore(layer, func(host string, id string, publish packet.Publish) error {
		session, err := broker.Sessions.ByID(id)
		if err != nil {
			log.Printf("ERR: session not found")
			return err
		}
		return session.Transport.Publish(broker.ctx, &publish)
	})
	if err != nil {
		log.Fatal(err)
	}
	peersStore := broker.mesh.Peers()
	topicssStore, err := topics.NewMemDBStore(layer)
	if err != nil {
		log.Fatal(err)
	}
	sessionsStore, err := sessions.NewSessionStore(layer, broker.RemoteRPCProvider, logrus.New().WithField("source", "session_store"))
	if err != nil {
		log.Fatal(err)
	}
	broker.Peers = peersStore
	broker.Topics = topicssStore
	broker.Subscriptions = subscriptionsStore
	broker.Sessions = sessionsStore
	broker.Peers.On(cluster.PeerDeleted, broker.onPeerDown)
}
func (b *Broker) onPeerDown(peer cluster.Peer) {
	name := peer.ID
	set, err := b.Subscriptions.ByPeer(name)
	if err != nil {
		log.Printf("ERR: failed to remove subscriptions from peer %s: %v", name, err)
		return
	}
	set.Apply(func(sub subscriptions.Subscription) {
		log.Printf("INFO: removing subscription %s", sub.ID)
		b.Subscriptions.Delete(sub.ID)
	})

	sessionSet, err := b.Sessions.ByPeer(name)
	if err != nil {
		log.Printf("ERR: failed to fetch sessions from peer %s: %v", name, err)
		return
	}
	sessionSet.Apply(func(s sessions.Session) {
		log.Printf("INFO: removing session %s", s.ID)
		b.Sessions.Delete(s.ID, "session_lost")
		if s.WillRetain {
			retainedMessage := topics.RetainedMessage{
				Metadata: topics.Metadata{
					Payload: s.WillPayload,
					Qos:     s.WillQoS,
					Tenant:  s.Tenant,
					Topic:   s.WillTopic,
				},
			}
			b.Topics.Create(retainedMessage)
		}
		recipients, err := b.Subscriptions.ByTopic(s.Tenant, s.WillTopic)
		if err != nil {
			return
		}

		lwt := packet.Publish{
			Payload: s.WillPayload,
			Topic:   s.WillTopic,
			Header: &packet.Header{
				Qos: s.WillQoS,
			},
		}
		recipients.Apply(func(sub subscriptions.Subscription) {
			sub.Sender(b.ctx, lwt)
		})
	})
}

func (b *Broker) Join(hosts []string) {
	b.mesh.Join(hosts)
}

func (b *Broker) isSessionLocal(session sessions.Session) bool {
	return session.Metadata.Peer == b.ID
}

func (b *Broker) dispatch(message *pb.MessagePublished) error {
	session, err := b.Sessions.ByID(message.Recipient)
	if err != nil {
		return err
	}
	packet := packet.Publish{
		Header: &packet.Header{
			Dup:    message.Dup,
			Qos:    message.Qos,
			Retain: message.Retained,
		},
		Payload:   message.Payload,
		Topic:     message.Topic,
		MessageId: 1,
	}
	return session.Transport.Publish(b.ctx, &packet)
}

func (b *Broker) Stop() {
	log.Printf("INFO: stopping broker")
	b.publishQueue.Close()
	log.Printf("INFO: broker stopped")
}
