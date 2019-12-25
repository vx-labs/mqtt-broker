package broker

import (
	"context"

	"github.com/vx-labs/mqtt-broker/transport"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/vx-labs/mqtt-broker/cluster"

	sessions "github.com/vx-labs/mqtt-broker/services/sessions/pb"
	topics "github.com/vx-labs/mqtt-broker/services/topics/pb"

	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
)

type QueuesStore interface {
	Create(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}
type SessionStore interface {
	RefreshKeepAlive(ctx context.Context, id string, timestamp int64) error
}

type TopicStore interface {
	ByTopicPattern(ctx context.Context, tenant string, pattern []byte) ([]*topics.RetainedMessage, error)
}
type MessagesStore interface {
	Put(ctx context.Context, streamId string, shardKey string, payload []byte) error
}
type Broker struct {
	ID         string
	logger     *zap.Logger
	authHelper func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error)
	mesh       cluster.Mesh
	Sessions   SessionStore
	Messages   *messages.Client
	grpcServer *grpc.Server
	ctx        context.Context
}

func New(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer, config Config) *Broker {
	ctx := context.Background()
	sessionsConn, err := mesh.DialService("sessions")
	if err != nil {
		panic(err)
	}
	broker := &Broker{
		ID:         id,
		authHelper: config.AuthHelper,
		ctx:        ctx,
		mesh:       mesh,
		logger:     logger,
		Sessions:   sessions.NewClient(sessionsConn),
	}

	return broker
}
