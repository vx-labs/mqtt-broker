package router

import (
	"context"

	"github.com/vx-labs/mqtt-broker/cluster"
	kv "github.com/vx-labs/mqtt-broker/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	queues "github.com/vx-labs/mqtt-broker/queues/pb"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions/pb"

	"go.uber.org/zap"
)

const (
	dbPath = "/var/tmp/mqtt-broker-messages"
)

type server struct {
	id            string
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *zap.Logger
	Subscriptions *subscriptions.Client
	Queues        *queues.Client
	Messages      *messages.Client
	KV            *kv.Client
	done          chan struct{}
}

func New(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) *server {
	ctx, cancel := context.WithCancel(context.Background())
	b := &server{
		id:     id,
		ctx:    ctx,
		logger: logger,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	kvConn, err := mesh.DialService("kv?tags=leader")
	if err != nil {
		panic(err)
	}
	messagesConn, err := mesh.DialService("messages")
	if err != nil {
		panic(err)
	}
	queuesConn, err := mesh.DialService("queues?tags=leader")
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
	//FIXME: routine leak
	go b.consumeMessages()

	return b
}
