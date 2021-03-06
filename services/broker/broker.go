package broker

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	topics "github.com/vx-labs/mqtt-broker/services/topics/pb"

	auth "github.com/vx-labs/mqtt-broker/services/auth/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
)

type QueuesStore interface {
	Create(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
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
	Messages   *messages.Client
	auth       *auth.Client
	grpcServer *grpc.Server
	listener   net.Listener
	ctx        context.Context
}

func New(id string, logger *zap.Logger, mesh discovery.DiscoveryAdapter) *Broker {
	ctx := context.Background()
	authConn, err := mesh.DialService("auth", "rpc")
	if err != nil {
		panic(err)
	}
	messagesConn, err := mesh.DialService("messages", "rpc")
	if err != nil {
		panic(err)
	}
	broker := &Broker{
		ID:       id,
		ctx:      ctx,
		logger:   logger,
		auth:     auth.NewClient(authConn),
		Messages: messages.NewClient(messagesConn),
	}

	return broker
}
