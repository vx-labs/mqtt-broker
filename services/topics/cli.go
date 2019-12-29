package topics

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"

	proto "github.com/golang/protobuf/proto"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-broker/network"
	broker "github.com/vx-labs/mqtt-broker/services/broker/pb"
	kv "github.com/vx-labs/mqtt-broker/services/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	queues "github.com/vx-labs/mqtt-broker/services/queues/pb"
	"github.com/vx-labs/mqtt-broker/stream"
	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/vx-labs/mqtt-broker/services/topics/pb"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	id         string
	store      *memDBStore
	state      types.GossipServiceLayer
	Queues     *queues.Client
	logger     *zap.Logger
	ctx        context.Context
	gprcServer *grpc.Server
	cancel     chan struct{}
	done       chan struct{}
}

func New(id string, logger *zap.Logger) *server {

	return &server{
		id:     id,
		logger: logger,
	}
}

func (b *server) Shutdown() {
	close(b.cancel)
	b.gprcServer.GracefulStop()
	<-b.done
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh discovery.DiscoveryAdapter) {
	l := cluster.NewGossipServiceLayer(name, logger, config, mesh)
	db, err := NewMemDBStore(l)
	if err != nil {
		panic(err)
	}
	b.store = db
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

	k := kv.NewClient(kvConn)
	m := messages.NewClient(messagesConn)
	b.Queues = queues.NewClient(queuesConn)
	streamClient := stream.NewClient(k, m, logger)

	ctx := context.Background()
	b.ctx = ctx
	b.cancel = make(chan struct{})
	b.done = make(chan struct{})

	go func() {
		defer close(b.done)
		streamClient.Consume(ctx, b.cancel, "messages", b.consumeStream,
			stream.WithConsumerID(b.id),
			stream.WithConsumerGroupID("topics"),
			stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
		)
	}()
	go func() {
		streamClient.Consume(ctx, b.cancel, "events", b.consumeEventStream,
			stream.WithConsumerID(b.id),
			stream.WithConsumerGroupID("topics"),
			stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
		)
	}()
}

func (b *server) consumeStream(messages []*messages.StoredMessage) (int, error) {
	if b.store == nil {
		return 0, errors.New("store not ready")
	}
	for idx := range messages {
		publish := &broker.MessagePublished{}
		err := proto.Unmarshal(messages[idx].Payload, publish)
		if err != nil {
			return idx, errors.Wrap(err, "failed to decode message for shard")
		}
		if publish.Publish.Header.Retain {
			err := b.store.Create(pb.RetainedMessage{
				Tenant:  publish.Tenant,
				Payload: publish.Publish.Payload,
				Qos:     publish.Publish.Header.Qos,
				Topic:   publish.Publish.Topic,
			})
			if err != nil {
				return idx, err
			}
		}
	}
	return len(messages), nil
}
func (b *server) Health() string {
	return "ok"
}

func (m *server) Serve(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterTopicsServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	m.gprcServer = s
	return lis
}

func (m *server) ByTopicPattern(ctx context.Context, input *pb.ByTopicPatternInput) (*pb.ByTopicPatternOutput, error) {
	if m.store == nil {
		return nil, errors.New("server not ready")
	}
	messages, err := m.store.ByTopicPattern(input.Tenant, input.Pattern)
	if err != nil {
		return nil, err
	}
	out := &pb.ByTopicPatternOutput{
		Messages: make([]*pb.RetainedMessage, len(messages)),
	}
	for idx := range messages {
		out.Messages[idx] = &messages[idx]
	}
	return out, nil
}

func (b *server) consumeEventStream(messages []*messages.StoredMessage) (int, error) {
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
			case *events.StateTransition_SessionSubscribed:
				input := event.SessionSubscribed
				messages, err := b.store.ByTopicPattern(input.Tenant, input.Pattern)
				if err != nil {
					b.logger.Error("failed to lookup store for retained messages")
					return idx, err
				}
				batches := make([]queues.MessageBatch, len(messages))
				for idx := range batches {
					batches[idx] = queues.MessageBatch{
						ID: input.SessionID,
						Publish: &packet.Publish{
							Header: &packet.Header{
								Retain: true,
							},
							Payload: messages[idx].Payload,
							Topic:   messages[idx].Topic,
						},
					}
				}
				err = b.Queues.PutMessageBatch(b.ctx, batches)
				if err != nil {
					b.logger.Error("failed to store retained messages in queue")
					return idx, err
				}
			}
		}
	}
	return len(messages), nil
}
