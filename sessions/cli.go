package sessions

import (
	"context"
	fmt "fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/events"
	kv "github.com/vx-labs/mqtt-broker/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/sessions/pb"
	"github.com/vx-labs/mqtt-broker/stream"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	b.state.Leave()
	b.gprcServer.GracefulStop()
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	l := cluster.NewGossipServiceLayer(name, logger, config, mesh)
	b.store = NewSessionStore(l, logger)
	b.state = l
	kvConn, err := mesh.DialService("kv?raft_status=leader")
	if err != nil {
		panic(err)
	}
	messagesConn, err := mesh.DialService("messages")
	if err != nil {
		panic(err)
	}
	k := kv.NewClient(kvConn)
	m := messages.NewClient(messagesConn)
	streamClient := stream.NewClient(k, m, logger)

	ctx := context.Background()
	b.cancel = make(chan struct{})
	b.done = make(chan struct{})

	go func() {
		defer close(b.done)
		streamClient.Consume(ctx, b.cancel, "events", b.consumeStream,
			stream.WithConsumerID(b.id),
			stream.WithConsumerGroupID("sessions"),
			stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
		)
	}()
}

func (b *server) consumeStream(messages []*messages.StoredMessage) (int, error) {
	if b.store == nil {
		return 0, errors.New("store not ready")
	}
	for idx := range messages {
		eventSet, err := events.Decode(messages[idx].Payload)
		if err != nil {
			return idx, errors.Wrap(err, "failed to decode message for shard")
		}
		for _, eventPayload := range eventSet {
			switch event := eventPayload.GetEvent().(type) {
			case *events.StateTransition_SessionClosed:
				input := event.SessionClosed
				err := b.store.Delete(input.ID)
				if err != nil {
					b.logger.Warn("failed to delete session", zap.Error(err))
				}
			case *events.StateTransition_SessionLost:
				input := event.SessionLost
				err := b.store.Delete(input.ID)
				if err != nil {
					b.logger.Warn("failed to delete session", zap.Error(err))
				}
			case *events.StateTransition_SessionCreated:
				input := event.SessionCreated
				err := b.store.Create(&pb.Session{
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
				})
				if err != nil {
					b.logger.Error("failed to create session", zap.Error(err))
					return idx, err
				}
			}
		}

	}
	return len(messages), nil
}

func (m *server) Health() string {
	if m.state == nil {
		return "warning"
	}
	return m.state.Health()
}
func (m *server) Serve(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterSessionsServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	m.gprcServer = s
	return lis
}
