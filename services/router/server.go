package router

import (
	"context"

	"github.com/vx-labs/mqtt-broker/stream"

	"github.com/vx-labs/mqtt-broker/cluster"
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
	done          chan struct{}
	cancel        chan struct{}
}

func New(id string, logger *zap.Logger, mesh cluster.DiscoveryAdapter) *server {
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

	streamClient := stream.NewClient(b.KV, b.Messages, logger)
	b.cancel = make(chan struct{})
	b.done = make(chan struct{})

	go func() {
		defer close(b.done)
		streamClient.Consume(ctx, b.cancel, "messages", b.v2ConsumePayload,
			stream.WithConsumerID(b.id),
			stream.WithConsumerGroupID("router"),
			stream.WithMaxBatchSize(200),
			stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
		)
	}()
	return b
}
