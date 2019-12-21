package sessions

import (
	"context"
	"time"

	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/sessions/pb"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type server struct {
	id         string
	store      SessionStore
	state      types.GossipServiceLayer
	ctx        context.Context
	gprcServer *grpc.Server
	logger     *zap.Logger
	cancel     chan struct{}
	done       chan struct{}
}

func New(id string, logger *zap.Logger) *server {

	return &server{
		id:     id,
		ctx:    context.Background(),
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
	return m.store.All(input)
}
func (m *server) RefreshKeepAlive(ctx context.Context, input *pb.RefreshKeepAliveInput) (*pb.RefreshKeepAliveOutput, error) {
	err := m.store.Update(input.ID, func(session pb.Session) *pb.Session {
		session.LastKeepAlive = input.Timestamp
		return &session
	})

	return &pb.RefreshKeepAliveOutput{}, err
}
func (m *server) Create(ctx context.Context, input *pb.SessionCreateInput) (*pb.SessionCreateOutput, error) {
	return &pb.SessionCreateOutput{}, nil
}
func (m *server) Delete(ctx context.Context, input *pb.SessionDeleteInput) (*pb.SessionDeleteOutput, error) {
	return &pb.SessionDeleteOutput{}, nil
}

func isSessionExpired(session *pb.Session, now int64) bool {
	return session.Created == 0 ||
		session.LastKeepAlive == 0 ||
		now-session.LastKeepAlive > 2*int64(session.KeepaliveInterval)
}
func (m *server) deleteSession(id string) error {
	return m.store.Delete(id)
}
func (m *server) gcExpiredSessions() {
	sessions, err := m.store.All(nil)
	if err != nil {
		m.logger.Error("failed to gc sessions", zap.Error(err))
	}
	now := time.Now().Unix()
	for _, session := range sessions.Sessions {
		if isSessionExpired(session, now) {
			err := m.deleteSession(session.ID)
			if err != nil {
				m.logger.Error("failed to gc session", zap.String("session_id", session.ID), zap.Error(err))
			} else {
				m.logger.Info("deleted expired session", zap.String("session_id", session.ID))
			}
		}
	}
}
