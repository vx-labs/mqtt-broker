package broker

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/vx-labs/mqtt-broker/cluster"

	sessions "github.com/vx-labs/mqtt-broker/services/sessions/pb"
	topics "github.com/vx-labs/mqtt-broker/services/topics/pb"

	auth "github.com/vx-labs/mqtt-broker/services/auth/pb"
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
	mesh       cluster.Mesh
	Sessions   SessionStore
	Messages   *messages.Client
	auth       *auth.Client
	grpcServer *grpc.Server
	ctx        context.Context
}

func New(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) *Broker {
	ctx := context.Background()
	sessionsConn, err := mesh.DialService("sessions")
	if err != nil {
		panic(err)
	}
	authConn, err := mesh.DialService("auth")
	if err != nil {
		panic(err)
	}
	messagesConn, err := mesh.DialService("messages?raft_status=leader")
	if err != nil {
		panic(err)
	}
	broker := &Broker{
		ID:       id,
		ctx:      ctx,
		mesh:     mesh,
		logger:   logger,
		Sessions: sessions.NewClient(sessionsConn),
		auth:     auth.NewClient(authConn),
		Messages: messages.NewClient(messagesConn),
	}

	return broker
}
