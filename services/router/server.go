package router

import (
	"context"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/stream"

	kv "github.com/vx-labs/mqtt-broker/services/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	queues "github.com/vx-labs/mqtt-broker/services/queues/pb"
	subscriptions "github.com/vx-labs/mqtt-broker/services/subscriptions/pb"

	"go.uber.org/zap"
)

const (
	dbPath = "/var/tmp/mqtt-broker-messages"
)

type server struct {
	id            string
	ctx           context.Context
	logger        *zap.Logger
	Subscriptions *subscriptions.Client
	Queues        *queues.Client
	Messages      *messages.Client
	KV            *kv.Client
	stream        *stream.Client
}

func New(id string, logger *zap.Logger, mesh discovery.DiscoveryAdapter) *server {
	ctx := context.Background()
	b := &server{
		id:     id,
		ctx:    ctx,
		logger: logger,
	}
	kvConn, err := mesh.DialService("kv?raft_status=leader")
	if err != nil {
		panic(err)
	}
	messagesConn, err := mesh.DialService("messages")
	if err != nil {
		panic(err)
	}
	queuesConn, err := mesh.DialService("queues?raft_status=leader")
	if err != nil {
		panic(err)
	}
	subscriptionConn, err := mesh.DialService("subscriptions")
	if err != nil {
		panic(err)
	}
	b.KV = kv.NewClient(kvConn)
	b.Queues = queues.NewClient(queuesConn)
	b.Messages = messages.NewClient(messagesConn)
	b.Subscriptions = subscriptions.NewClient(subscriptionConn)

	b.stream = stream.NewClient(b.KV, b.Messages, logger)
	return b
}
