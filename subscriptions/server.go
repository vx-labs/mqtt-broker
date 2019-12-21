package subscriptions

import (
	"context"

	"github.com/vx-labs/mqtt-broker/cluster/types"
	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
	"github.com/vx-labs/mqtt-broker/subscriptions/pb"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type SessionStore interface {
	ByID(ctx context.Context, id string) (*sessions.Session, error)
}

type server struct {
	id         string
	store      Store
	state      types.GossipServiceLayer
	ctx        context.Context
	gprcServer *grpc.Server
	logger     *zap.Logger
	sessions   SessionStore
	cancel     chan struct{}
	done       chan struct{}
}

func New(id string, logger *zap.Logger) *server {

	return &server{
		id:     id,
		ctx:    context.Background(),
		logger: logger,
		cancel: make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (m *server) ByID(ctx context.Context, input *pb.SubscriptionByIDInput) (*pb.Subscription, error) {
	return m.store.ByID(input.ID)
}
func (m *server) BySession(ctx context.Context, input *pb.SubscriptionBySessionInput) (*pb.SubscriptionMetadataList, error) {
	return m.store.BySession(input.SessionID)
}
func (m *server) ByTopic(ctx context.Context, input *pb.SubscriptionByTopicInput) (*pb.SubscriptionMetadataList, error) {
	return m.store.ByTopic(input.Tenant, input.Topic)
}
func (m *server) All(ctx context.Context, input *pb.SubscriptionFilterInput) (*pb.SubscriptionMetadataList, error) {
	return m.store.All()
}
