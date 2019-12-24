package sessions

import (
	"context"

	"github.com/vx-labs/mqtt-broker/cluster/types"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/services/sessions/pb"
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
	Messages   *messages.Client
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
