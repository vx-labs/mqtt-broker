package subscriptions

import (
	"context"
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/subscriptions/pb"
	"go.uber.org/zap"
)

type server struct {
	id        string
	store     Store
	state     types.RaftServiceLayer
	ctx       context.Context
	listeners []net.Listener
	logger    *zap.Logger
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
	m.logger.Debug("deleting subscription", zap.String("subscription_id", input.ID))
	ev := pb.SubscriptionStateTransition{
		Kind: transitionSessionDeleted,
		SubscriptionDeleted: &pb.SubscriptionStateTransitionSubscriptionDeleted{
			ID: input.ID,
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return nil, err
	}
	return &pb.SubscriptionDeleteOutput{}, m.state.ApplyEvent(payload)
}
