package sessions

import (
	"context"
	"net"
	"time"

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
func (m *server) RefreshKeepAlive(ctx context.Context, input *pb.RefreshKeepAliveInput) (*pb.RefreshKeepAliveOutput, error) {
	session, err := m.store.ByID(input.ID)
	if err != nil {
		return nil, err
	}
	copy := *session
	copy.LastKeepAlive = input.Timestamp
	ev := pb.SessionStateTransition{
		Kind: transitionSessionCreated,
		SessionCreated: &pb.SessionStateTransitionSessionCreated{
			Input: &copy,
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return nil, err
	}
	return &pb.RefreshKeepAliveOutput{}, m.state.ApplyEvent(payload)
}
func (m *server) Create(ctx context.Context, input *pb.SessionCreateInput) (*pb.SessionCreateOutput, error) {
	m.logger.Debug("creating session", zap.String("session_id", input.ID))
	ev := pb.SessionStateTransition{
		Kind: transitionSessionCreated,
		SessionCreated: &pb.SessionStateTransitionSessionCreated{
			Input: &pb.Session{
				ClientID:          input.ClientID,
				ID:                input.ID,
				KeepaliveInterval: input.KeepaliveInterval,
				Peer:              input.Peer,
				RemoteAddress:     input.RemoteAddress,
				Tenant:            input.Tenant,
				Transport:         input.Transport,
				WillPayload:       input.WillPayload,
				WillTopic:         input.WillTopic,
				WillRetain:        input.WillRetain,
				WillQoS:           input.WillQoS,
				Created:           input.Timestamp,
				LastKeepAlive:     input.Timestamp,
			},
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return nil, err
	}
	return &pb.SessionCreateOutput{}, m.state.ApplyEvent(payload)
}
func (m *server) Delete(ctx context.Context, input *pb.SessionDeleteInput) (*pb.SessionDeleteOutput, error) {
	return &pb.SessionDeleteOutput{}, m.deleteSession(input.ID)
}

func isSessionExpired(session *pb.Session, now int64) bool {
	return session.Created == 0 ||
		session.LastKeepAlive == 0 ||
		now-session.LastKeepAlive > 2*int64(session.KeepaliveInterval)
}
func (m *server) deleteSession(id string) error {
	m.logger.Debug("deleting session", zap.String("session_id", id))
	ev := pb.SessionStateTransition{
		Kind: transitionSessionDeleted,
		SessionDeleted: &pb.SessionStateTransitionSessionDeleted{
			ID: id,
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return err
	}
	return m.state.ApplyEvent(payload)
}
func (m *server) gcExpiredSessions() {
	if !m.state.IsLeader() {
		return
	}
	sessions, err := m.store.All()
	if err != nil {
		m.logger.Error("failed to gc sessions", zap.Error(err))
	}
	now := time.Now().Unix()
	for _, session := range sessions.Sessions {
		if isSessionExpired(session, now) {
			err := m.deleteSession(session.ID)
			if err != nil {
				m.logger.Error("failed to gc session", zap.String("session_id", session.ID), zap.Error(err))
			}
		}
	}
}
