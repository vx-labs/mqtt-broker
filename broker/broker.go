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

	queues "github.com/vx-labs/mqtt-broker/queues/pb"
	publishQueue "github.com/vx-labs/mqtt-broker/struct/queues/publish"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions/pb"
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
type QueuesStore interface {
	Create(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	PutMessage(ctx context.Context, id string, publish *packet.Publish) error
}
type SessionStore interface {
	ByID(ctx context.Context, id string) (*sessions.Session, error)
	ByClientID(ctx context.Context, id string) ([]*sessions.Session, error)
	ByPeer(ctx context.Context, peer string) ([]*sessions.Session, error)
	All(ctx context.Context) ([]*sessions.Session, error)
	Create(ctx context.Context, sess sessions.SessionCreateInput) error
	RefreshKeepAlive(ctx context.Context, id string, timestamp int64) error
	Delete(ctx context.Context, id string) error
}

type TopicStore interface {
	Create(message topics.RetainedMessage) error
	ByTopicPattern(tenant string, pattern []byte) (topics.RetainedMessageSet, error)
	All() (topics.RetainedMessageSet, error)
}
type SubscriptionStore interface {
	ByTopic(ctx context.Context, tenant string, pattern []byte) ([]*subscriptions.Metadata, error)
	ByID(ctx context.Context, id string) (*subscriptions.Metadata, error)
	All(ctx context.Context) ([]*subscriptions.Metadata, error)
	ByPeer(ctx context.Context, peer string) ([]*subscriptions.Metadata, error)
	BySession(ctx context.Context, id string) ([]*subscriptions.Metadata, error)
	Create(ctx context.Context, message subscriptions.SubscriptionCreateInput) error
	Delete(ctx context.Context, id string) error
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
	Queues        QueuesStore
	Peers         PeerStore
	workers       *pool.Pool
	ctx           context.Context
	publishQueue  Queue
}

func New(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer, config Config) *Broker {
	ctx := context.Background()
	sessionsConn, err := mesh.DialService("sessions?tags=leader")
	if err != nil {
		panic(err)
	}
	subscriptionsConn, err := mesh.DialService("subscriptions?tags=leader")
	if err != nil {
		panic(err)
	}
	queuesConn, err := mesh.DialService("queues")
	if err != nil {
		panic(err)
	}
	broker := &Broker{
		ID:            id,
		authHelper:    config.AuthHelper,
		workers:       pool.NewPool(25),
		ctx:           ctx,
		mesh:          mesh,
		Queues:        queues.NewClient(queuesConn),
		publishQueue:  publishQueue.New(),
		logger:        logger,
		Sessions:      sessions.NewClient(sessionsConn),
		Subscriptions: subscriptions.NewClient(subscriptionsConn),
	}

	broker.startPublishConsumers()
	return broker
}

func (broker *Broker) Start(layer types.GossipServiceLayer) {
	peersStore := broker.mesh.Peers()
	topicssStore, err := topics.NewMemDBStore(layer)
	if err != nil {
		log.Fatal(err)
	}
	broker.Peers = peersStore
	broker.Topics = topicssStore
	layer.OnNodeLeave(func(id string, meta clusterpb.NodeMeta) {
		broker.onPeerDown(id)
	})
}
func (b *Broker) onPeerDown(name string) {
	set, err := b.Subscriptions.ByPeer(b.ctx, name)
	if err != nil {
		b.logger.Error("failed to remove subscriptions from old peer", zap.String("peer_id", name), zap.Error(err))
		return
	}
	if len(set) > 0 {
		for _, sub := range set {
			b.Subscriptions.Delete(b.ctx, sub.ID)
		}
		b.logger.Info("removed subscriptions from old peer", zap.String("peer_id", name), zap.Int("count", len(set)))
	}
	sessionSet, err := b.Sessions.ByPeer(b.ctx, name)
	if err != nil {
		b.logger.Error("failed to remove sessions from old peer", zap.String("peer_id", name), zap.Error(err))
		return
	}
	for _, s := range sessionSet {
		b.Sessions.Delete(b.ctx, s.ID)
		if len(s.WillTopic) > 0 {
			lwt := &packet.Publish{
				Payload: s.WillPayload,
				Topic:   s.WillTopic,
				Header: &packet.Header{
					Qos:    s.WillQoS,
					Retain: s.WillRetain,
				},
			}
			b.publishQueue.Enqueue(&publishQueue.Message{
				Tenant:  s.Tenant,
				Publish: lwt,
			})
		}
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
