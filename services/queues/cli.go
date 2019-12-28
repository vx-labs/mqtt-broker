package queues

import (
	"context"
	fmt "fmt"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/network"
	kv "github.com/vx-labs/mqtt-broker/services/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/services/queues/pb"
	sessions "github.com/vx-labs/mqtt-broker/services/sessions/pb"
	"github.com/vx-labs/mqtt-broker/stream"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	close(b.cancel)
	<-b.done
	err := b.state.Shutdown()
	if err != nil {
		b.logger.Error("failed to shutdown raft instance cleanly", zap.Error(err))
	}
	b.gprcServer.GracefulStop()
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryAdapter) {
	var err error
	sessionConn, err := mesh.DialService("sessions")
	if err != nil {
		logger.Fatal("failed to dial session service")
	}
	b.sessions = sessions.NewClient(sessionConn)
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	err = b.state.Start(name, b)
	if err != nil {
		panic(err)
	}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			if b.state.IsLeader() {
				expiredInflights := b.store.GetExpiredInflights(time.Now())
				if len(expiredInflights) > 0 {
					err := b.commitEvent(&pb.QueuesStateTransition{
						Kind: MessageInflightExpired,
						MessageInflightExpired: &pb.QueueStateTransitionMessageInflightExpired{
							ExpiredInflights: expiredInflights,
						},
					})
					if err != nil {
						b.logger.Error("failed to commit expired inflight message", zap.Error(err))
					}
				}
			}
		}
	}()
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
	b.Messages = m
	ctx := context.Background()
	b.cancel = make(chan struct{})
	b.done = make(chan struct{})

	go func() {
		defer close(b.done)
		streamClient.Consume(ctx, b.cancel, "events", b.consumeStream,
			stream.WithConsumerID(b.id),
			stream.WithConsumerGroupID("queues"),
			stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
		)
	}()

}
func (m *server) Health() string {
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
	pb.RegisterQueuesServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	m.gprcServer = s
	return lis
}
