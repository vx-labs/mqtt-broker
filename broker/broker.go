package broker

import (
	"context"
	"log"

	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/pool"
	"github.com/vx-labs/mqtt-broker/transport"
	"go.uber.org/zap"

	"github.com/vx-labs/mqtt-broker/cluster"
	clusterpb "github.com/vx-labs/mqtt-broker/cluster/pb"

	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
	"github.com/vx-labs/mqtt-broker/topics"
	topicspb "github.com/vx-labs/mqtt-broker/topics/pb"

	"github.com/vx-labs/mqtt-protocol/packet"

	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	queues "github.com/vx-labs/mqtt-broker/queues/pb"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions/pb"
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
	PutMessageBatch(ctx context.Context, batch []queues.MessageBatch) error
}
type SessionStore interface {
	ByID(ctx context.Context, id string) (*sessions.Session, error)
	ByClientID(ctx context.Context, id string) ([]*sessions.Session, error)
	ByPeer(ctx context.Context, peer string) ([]*sessions.Session, error)
	Create(ctx context.Context, sess sessions.SessionCreateInput) error
	RefreshKeepAlive(ctx context.Context, id string, timestamp int64) error
	Delete(ctx context.Context, id string) error
}

type TopicStore interface {
	Create(message topicspb.RetainedMessage) error
	ByTopicPattern(tenant string, pattern []byte) (topicspb.RetainedMessageSet, error)
	All() (topicspb.RetainedMessageSet, error)
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
type MessagesStore interface {
	Put(ctx context.Context, streamId string, shardKey string, payload []byte) error
	CreateStream(ctx context.Context, streamId string, shardCount int) error
	GetStream(ctx context.Context, streamId string) (*messages.StreamConfig, error)
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
	Messages      MessagesStore
	Peers         PeerStore
	workers       *pool.Pool
	ctx           context.Context
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
	queuesConn, err := mesh.DialService("queues?tags=leader")
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
		logger:        logger,
		Sessions:      sessions.NewClient(sessionsConn),
		Subscriptions: subscriptions.NewClient(subscriptionsConn),
	}

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
			b.enqueuePublish(s.Tenant, lwt)
		}
	}
}

func (b *Broker) Stop() {
}
