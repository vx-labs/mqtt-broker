package queues

import (
	"context"
	fmt "fmt"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/adapters/cp"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
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
	b.stream.Shutdown()
	err := b.state.Shutdown()
	if err != nil {
		b.logger.Error("failed to shutdown raft instance cleanly", zap.Error(err))
	}
	b.gprcServer.GracefulStop()
}
func (b *server) Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error {
	userService := catalog.Service(name)
	raftService := catalog.Service(fmt.Sprintf("%sgossip", name))
	raftRPCService := catalog.Service(fmt.Sprintf("%sgossiprpc", name))
	listener, err := userService.ListenTCP()
	if err != nil {
		return err
	}
	b.listener = listener
	b.state = cp.RaftSynchronizer(id, userService, raftService, raftRPCService, b, logger)

	sessionConn, err := catalog.Dial("sessions")
	if err != nil {
		logger.Error("failed to dial session service")
		return err
	}
	b.sessions = sessions.NewClient(sessionConn)

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
	kvConn, err := catalog.Dial("kv")
	if err != nil {
		panic(err)
	}
	messagesConn, err := catalog.Dial("messages")
	if err != nil {
		panic(err)
	}
	k := kv.NewClient(kvConn)
	m := messages.NewClient(messagesConn)
	b.stream = stream.NewClient(k, m, logger)
	b.Messages = m
	ctx := context.Background()

	b.stream.ConsumeStream(ctx, "events", b.consumeStream,
		stream.WithConsumerID(b.id),
		stream.WithConsumerGroupID("queues"),
		stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
	)
	return nil
}
func (m *server) Health() string {
	if m.state == nil {
		return "warning"
	}
	return m.state.Health()
}
func (m *server) Serve(port int) net.Listener {
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterQueuesServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(m.listener)
	m.gprcServer = s
	return m.listener
}
