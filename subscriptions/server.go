package subscriptions

import (
	"context"

	"github.com/gogo/protobuf/proto"
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
	state      types.RaftServiceLayer
	ctx        context.Context
	gprcServer *grpc.Server
	logger     *zap.Logger
	sessions   SessionStore
}

func New(id string, logger *zap.Logger) *server {

	return &server{
		id:     id,
		ctx:    context.Background(),
		store:  NewMemDBStore(),
		logger: logger,
	}
}

func (m *server) ByID(ctx context.Context, input *pb.SubscriptionByIDInput) (*pb.Metadata, error) {
	return m.store.ByID(input.ID)
}
func (m *server) BySession(ctx context.Context, input *pb.SubscriptionBySessionInput) (*pb.SubscriptionMetadataList, error) {
	return m.store.BySession(input.SessionID)
}
func (m *server) ByTopic(ctx context.Context, input *pb.SubscriptionByTopicInput) (*pb.SubscriptionMetadataList, error) {
	return m.store.ByTopic(input.Tenant, input.Topic)
}
func (m *server) ByPeer(ctx context.Context, input *pb.SubscriptionByPeerInput) (*pb.SubscriptionMetadataList, error) {
	return m.store.ByPeer(input.PeerID)
}
func (m *server) All(ctx context.Context, input *pb.SubscriptionFilterInput) (*pb.SubscriptionMetadataList, error) {
	return m.store.All()
}
func (m *server) Create(ctx context.Context, input *pb.SubscriptionCreateInput) (*pb.SubscriptionCreateOutput, error) {
	m.logger.Debug("creating subscription", zap.String("subscription_id", input.ID))
	ev := pb.SubscriptionStateTransition{
		Kind: transitionSessionCreated,
		SubscriptionCreated: &pb.SubscriptionStateTransitionSubscriptionCreated{
			Input: input,
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return nil, err
	}
	return &pb.SubscriptionCreateOutput{}, m.state.ApplyEvent(payload)
}
func (m *server) Delete(ctx context.Context, input *pb.SubscriptionDeleteInput) (*pb.SubscriptionDeleteOutput, error) {
	return &pb.SubscriptionDeleteOutput{}, m.deleteSubscription(input.ID)
}
func (m *server) deleteSubscription(id string) error {
	m.logger.Debug("deleting subscription", zap.String("subscription_id", id))
	ev := pb.SubscriptionStateTransition{
		Kind: transitionSessionDeleted,
		SubscriptionDeleted: &pb.SubscriptionStateTransitionSubscriptionDeleted{
			ID: id,
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return err
	}
	return m.state.ApplyEvent(payload)
}

func (m *server) isSubscriptionExpired(sub *pb.Metadata) bool {
	_, err := m.sessions.ByID(m.ctx, sub.SessionID)
	return err != nil
}
func (m *server) gcExpiredSubscriptions() {
	if !m.state.IsLeader() {
		return
	}
	subscriptions, err := m.store.All()
	if err != nil {
		m.logger.Error("failed to gc subscriptions", zap.Error(err))
	}
	for _, subscription := range subscriptions.Metadatas {
		if m.isSubscriptionExpired(subscription) {
			err := m.deleteSubscription(subscription.ID)
			if err != nil {
				m.logger.Error("failed to gc subscription", zap.String("subscription_id", subscription.ID), zap.Error(err))
			} else {
				m.logger.Info("deleted expired subscription", zap.String("subscription_id", subscription.ID))
			}
		}
	}
}
