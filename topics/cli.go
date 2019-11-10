package topics

import (
	"context"
	"errors"
	"fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/network"
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
