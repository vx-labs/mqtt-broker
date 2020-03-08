package sessions

import (
	"context"
	"net"

	ap "github.com/vx-labs/mqtt-broker/adapters/ap"
	"github.com/vx-labs/mqtt-broker/events"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/services/sessions/pb"
	"github.com/vx-labs/mqtt-broker/stream"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type server struct {
	id         string
	store      SessionStore
	state      ap.Distributer
	ctx        context.Context
	gprcServer *grpc.Server
	listener   net.Listener
	logger     *zap.Logger
	Messages   *messages.Client
	stream     *stream.Client
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
	session, err := m.store.ByID(input.ID)
	if err != nil {
		return nil, err
	}
	err = events.Commit(ctx, m.Messages, input.ID, &events.StateTransition{
		Event: &events.StateTransition_SessionKeepalived{
			SessionKeepalived: &events.SessionKeepalived{
				SessionID: input.ID,
				Tenant:    session.Tenant,
				Timestamp: input.Timestamp,
			},
		},
	})
	return &pb.RefreshKeepAliveOutput{}, err
}
func (m *server) Create(ctx context.Context, input *pb.SessionCreateInput) (*pb.SessionCreateOutput, error) {
	return &pb.SessionCreateOutput{}, nil
}
func (m *server) Delete(ctx context.Context, input *pb.SessionDeleteInput) (*pb.SessionDeleteOutput, error) {
	session, err := m.store.ByID(input.ID)
	if err != nil {
		return nil, err
	}
	err = events.Commit(ctx, m.Messages, input.ID, &events.StateTransition{
		Event: &events.StateTransition_SessionLost{
			SessionLost: &events.SessionLost{
				ID:     input.ID,
				Tenant: session.Tenant,
			},
		},
	})
	return &pb.SessionDeleteOutput{}, err
}

func (m *server) deleteSession(id string) error {
	return m.store.Delete(id)
}
