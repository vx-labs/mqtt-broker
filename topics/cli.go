package topics

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"

	proto "github.com/golang/protobuf/proto"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	broker "github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	kv "github.com/vx-labs/mqtt-broker/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/messages/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/stream"
	"github.com/vx-labs/mqtt-broker/topics/pb"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	id     string
	store  *memDBStore
	state  types.GossipServiceLayer
	logger *zap.Logger
}

func New(id string, logger *zap.Logger) *server {

	return &server{
		id:     id,
		logger: logger,
	}
}

func (b *server) Shutdown() {
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
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
	k := kv.NewClient(kvConn)
	m := messages.NewClient(messagesConn)
	streamClient := stream.NewClient(k, m, logger)

	ctx := context.Background()
	logger.Debug("started stream consumer")
	//FIXME: routine leak
	go streamClient.Consume(ctx, "messages", b.consumeStream,
		stream.WithConsumerID(b.id),
		stream.WithConsumerGroupID("topics"),
		stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
	)
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
