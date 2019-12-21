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
}

func New(id string, logger *zap.Logger) *server {

	return &server{
		id:     id,
		ctx:    context.Background(),
		logger: logger,
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
func (m *server) Create(ctx context.Context, input *pb.SubscriptionCreateInput) (*pb.SubscriptionCreateOutput, error) {
	m.logger.Debug("creating subscription", zap.String("subscription_id", input.ID))
	return &pb.SubscriptionCreateOutput{}, m.store.Create(&pb.Subscription{
		ID:        input.ID,
		Pattern:   input.Pattern,
		Qos:       input.Qos,
		SessionID: input.SessionID,
		Tenant:    input.Tenant,
	})
}
func (m *server) Delete(ctx context.Context, input *pb.SubscriptionDeleteInput) (*pb.SubscriptionDeleteOutput, error) {
	return &pb.SubscriptionDeleteOutput{}, m.deleteSubscription(input.ID)
}
func (m *server) deleteSubscription(id string) error {
	m.logger.Debug("deleting subscription", zap.String("subscription_id", id))
	return m.store.Delete(id)
}
