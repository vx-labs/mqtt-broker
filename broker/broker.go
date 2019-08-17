package broker

import (
	"context"
	"log"

	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/pool"
	"github.com/vx-labs/mqtt-broker/transport"
	"go.uber.org/zap"

	"github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/cluster"
	clusterpb "github.com/vx-labs/mqtt-broker/cluster/pb"

	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
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
	ByID(id string) (peers.Peer, error)
	All() (peers.SubscriptionSet, error)
	Delete(id string) error
	On(event string, handler func(peers.Peer)) func()
}
type SessionStore interface {
	ByID(ctx context.Context, id string) (*sessions.Session, error)
	ByClientID(ctx context.Context, id string) ([]*sessions.Session, error)
	ByPeer(ctx context.Context, peer string) ([]*sessions.Session, error)
	All(ctx context.Context) ([]*sessions.Session, error)
	Create(ctx context.Context, sess sessions.SessionCreateInput) error
	Delete(ctx context.Context, id string) error
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
	logger        *zap.Logger
	authHelper    func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error)
	mesh          cluster.Mesh
	Subscriptions SubscriptionStore
	Sessions      SessionStore
	Topics        TopicStore
	Peers         PeerStore
	workers       *pool.Pool
	ctx           context.Context
	publishQueue  Queue
}

func New(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer, config Config) *Broker {
	ctx := context.Background()
	sessionsConn, err := mesh.DialService("sessions")
	if err != nil {
		panic(err)
	}
	broker := &Broker{
		ID:           id,
		authHelper:   config.AuthHelper,
		workers:      pool.NewPool(25),
		ctx:          ctx,
		mesh:         mesh,
		publishQueue: publishQueue.New(),
		logger:       logger,
		Sessions:     sessions.NewClient(sessionsConn),
	}

	broker.startPublishConsumers()
	return broker
}

func (broker *Broker) Start(layer types.GossipServiceLayer) {
	subscriptionsStore, err := subscriptions.NewMemDBStore(layer, func(host string, id string, publish packet.Publish) error {
		session, err := broker.Sessions.ByID(broker.ctx, id)
		if err != nil {
			broker.logger.Warn("publish subscription toward an unknown session", zap.String("session_id", id), zap.Binary("topic_pattern", publish.Topic))
			set, err := broker.Subscriptions.BySession(id)
			if err != nil {
				broker.logger.Error("failed to fetch session subscriptions", zap.String("session_id", id), zap.Error(err))
				return err
			}
			set.Apply(func(subscription subscriptions.Subscription) {
				broker.Subscriptions.Delete(subscription.ID)
			})
			return nil
		}
		return broker.sendToSession(broker.ctx, id, session.Peer, &publish)
	})
	if err != nil {
		log.Fatal(err)
	}
	peersStore := broker.mesh.Peers()
	topicssStore, err := topics.NewMemDBStore(layer)
	if err != nil {
		log.Fatal(err)
	}
	broker.Peers = peersStore
	broker.Topics = topicssStore
	broker.Subscriptions = subscriptionsStore
	layer.OnNodeLeave(func(id string, meta clusterpb.NodeMeta) {
		broker.onPeerDown(id)
	})
}
func (b *Broker) onPeerDown(name string) {
	b.logger.Info("peer down", zap.String("peer_id", name))
	set, err := b.Subscriptions.ByPeer(name)
	if err != nil {
		b.logger.Error("failed to remove subscriptions from old peer", zap.String("peer_id", name), zap.Error(err))
		return
	}
	set.Apply(func(sub subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})
	b.logger.Info("removed subscriptions from old peer", zap.String("peer_id", name), zap.Int("count", len(set)))

	sessionSet, err := b.Sessions.ByPeer(b.ctx, name)
	if err != nil {
		b.logger.Error("failed to remove sessions from old peer", zap.String("peer_id", name), zap.Error(err))
		return
	}
	for _, s := range sessionSet {
		b.Sessions.Delete(b.ctx, s.ID)
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
	}
}

func (b *Broker) dispatch(message *pb.MessagePublished) error {
	session, err := b.Sessions.ByID(b.ctx, message.Recipient)
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
	return b.sendToSession(b.ctx, session.ID, session.Peer, &packet)
}

func (b *Broker) Stop() {
	b.publishQueue.Close()
}
