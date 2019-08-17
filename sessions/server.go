package sessions

import (
	"context"
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/sessions/pb"
	"go.uber.org/zap"
)

type server struct {
	id        string
	store     SessionStore
	state     types.RaftServiceLayer
	ctx       context.Context
	listeners []net.Listener
	logger    *zap.Logger
}

func New(id string, logger *zap.Logger) *server {

	return &server{
		id:     id,
		ctx:    context.Background(),
		store:  NewSessionStore(logger),
		logger: logger,
	}
}

func (m *server) ByID(ctx context.Context, input *pb.SessionByIDInput) (*pb.Session, error) {
	return m.store.ByID(input.ID)
}
func (m *server) ByClientID(ctx context.Context, input *pb.SessionByClientIDInput) (*pb.SessionMetadataList, error) {
	return m.store.ByClientID(input.ClientID)
}
func (m *server) ByPeer(ctx context.Context, input *pb.SessionByPeerInput) (*pb.SessionMetadataList, error) {
	return m.store.ByPeer(input.Peer)
}
func (m *server) All(ctx context.Context, input *pb.SessionFilterInput) (*pb.SessionMetadataList, error) {
	return m.store.All()
}
func (m *server) Create(ctx context.Context, input *pb.SessionCreateInput) (*pb.SessionCreateOutput, error) {
	m.logger.Debug("creating session", zap.String("session_id", input.ID))
	ev := pb.SessionStateTransition{
		Kind: transitionSessionCreated,
		SessionCreated: &pb.SessionStateTransitionSessionCreated{
			Input: input,
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return nil, err
	}
	return &pb.SessionCreateOutput{}, m.state.ApplyEvent(payload)
}
func (m *server) Delete(ctx context.Context, input *pb.SessionDeleteInput) (*pb.SessionDeleteOutput, error) {
	m.logger.Debug("deleting session", zap.String("session_id", input.ID))
	ev := pb.SessionStateTransition{
		Kind: transitionSessionDeleted,
		SessionDeleted: &pb.SessionStateTransitionSessionDeleted{
			ID: input.ID,
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return nil, err
	}
	return &pb.SessionDeleteOutput{}, m.state.ApplyEvent(payload)
}
