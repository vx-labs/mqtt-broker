package broker

import (
	"context"

	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/pool"
	"github.com/vx-labs/mqtt-broker/transport"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/vx-labs/mqtt-broker/cluster"

	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
	topics "github.com/vx-labs/mqtt-broker/topics/pb"

	"github.com/vx-labs/mqtt-protocol/packet"

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
	ByTopicPattern(ctx context.Context, tenant string, pattern []byte) ([]*topics.RetainedMessage, error)
}
type SubscriptionStore interface {
	ByTopic(ctx context.Context, tenant string, pattern []byte) ([]*subscriptions.Metadata, error)
	ByID(ctx context.Context, id string) (*subscriptions.Metadata, error)
	All(ctx context.Context) ([]*subscriptions.Metadata, error)
	BySession(ctx context.Context, id string) ([]*subscriptions.Metadata, error)
	Create(ctx context.Context, message subscriptions.SubscriptionCreateInput) error
	Delete(ctx context.Context, id string) error
}
type MessagesStore interface {
	Put(ctx context.Context, streamId string, shardKey string, payload []byte) error
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
	workers       *pool.Pool
	grpcServer    *grpc.Server
	ctx           context.Context
}

func New(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer, config Config) *Broker {
	ctx := context.Background()
	sessionsConn, err := mesh.DialService("sessions")
	if err != nil {
		panic(err)
	}
	subscriptionsConn, err := mesh.DialService("subscriptions?raft_status=leader")
	if err != nil {
		panic(err)
	}
	queuesConn, err := mesh.DialService("queues?raft_status=leader")
	if err != nil {
		panic(err)
	}
	topicsConn, err := mesh.DialService("topics")
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
		Topics:        topics.NewClient(topicsConn),
	}

	return broker
}
