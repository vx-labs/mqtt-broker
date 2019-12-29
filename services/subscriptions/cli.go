package subscriptions

import (
	"bytes"
	"context"
	"crypto/sha1"
	fmt "fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-broker/network"
	kv "github.com/vx-labs/mqtt-broker/services/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/services/subscriptions/pb"
	"github.com/vx-labs/mqtt-broker/stream"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

const (
	transitionSessionCreated = "session_created"
	transitionSessionDeleted = "session_deleted"
)

func (b *server) Shutdown() {
	b.state.Leave()
	b.gprcServer.GracefulStop()
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh discovery.DiscoveryAdapter) {
	l := cluster.NewGossipServiceLayer(name, logger, config, mesh)
	b.state = l
	b.store = NewSubscriptionStore(l, logger)
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
			stream.WithConsumerGroupID("subscriptions"),
			stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
		)
	}()
}
func makeSubID(session string, pattern []byte) string {
	hash := sha1.New()
	_, err := hash.Write([]byte(session))
	if err != nil {
		return ""
	}
	_, err = hash.Write(pattern)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
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
				set, err := b.store.BySession(input.ID)
				if err != nil {
					continue
				}
				for _, sub := range set.Subscriptions {
					err := b.store.Delete(sub.ID)
					if err != nil {
						b.logger.Warn("failed to delete subscription", zap.Error(err))
					}
				}
			case *events.StateTransition_SessionLost:
				input := event.SessionLost
				set, err := b.store.BySession(input.ID)
				if err != nil {
					continue
				}
				for _, sub := range set.Subscriptions {
					err := b.store.Delete(sub.ID)
					if err != nil {
						b.logger.Warn("failed to delete subscription", zap.Error(err))
					}
				}
			case *events.StateTransition_SessionSubscribed:
				input := event.SessionSubscribed
				id := makeSubID(input.SessionID, input.Pattern)
				err := b.store.Create(&pb.Subscription{
					ID:        id,
					Pattern:   input.Pattern,
					Qos:       input.Qos,
					SessionID: input.SessionID,
					Tenant:    input.Tenant,
				})
				if err != nil {
					b.logger.Error("failed to create subscription", zap.Error(err))
					return idx, err
				}
				b.logger.Info("session subscription created",
					zap.String("session_id", input.SessionID),
					zap.String("subscription_id", id),
					zap.Int32("qos", input.Qos),
					zap.Binary("topic_pattern", input.Pattern))

			case *events.StateTransition_SessionUnsubscribed:
				input := event.SessionUnsubscribed
				set, err := b.store.BySession(input.SessionID)
				if err != nil {
					continue
				}
				for _, sub := range set.Subscriptions {
					if bytes.Compare(sub.Pattern, input.Pattern) == 0 {
						err := b.store.Delete(sub.ID)
						if err != nil {
							return idx, err
						}
					}
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
	pb.RegisterSubscriptionsServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	m.gprcServer = s
	return lis
}
