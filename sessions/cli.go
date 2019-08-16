package sessions

import (
	"errors"
	fmt "fmt"
	"io"
	"net"

	"github.com/golang/protobuf/proto"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/sessions/pb"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

const (
	transitionSessionCreated = "session_created"
	transitionSessionDeleted = "session_deleted"
)

func (b *server) Shutdown() {
	for _, lis := range b.listeners {
		lis.Close()
	}
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	err := b.state.Start(name, b)
	if err != nil {
		panic(err)
	}
}
func (m *server) Restore(io.Reader) error {
	return nil
}
func (m *server) Snapshot() io.Reader {
	return nil
}
func (m *server) Health() string {
	return "ok"
}
func (m *server) Serve(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	pb.RegisterSessionsServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	return lis
}

func (m *server) Apply(payload []byte, leader bool) error {
	event := pb.SessionStateTransition{}
	err := proto.Unmarshal(payload, &event)
	if err != nil {
		return err
	}
	switch event.Kind {
	case transitionSessionCreated:
		m.logger.Info("created session", zap.String("session_id", event.SessionCreated.Input.ID))
		input := event.SessionCreated.Input
		err := m.store.Create(&pb.Session{
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
		})
		return err
	case transitionSessionDeleted:
		m.logger.Info("deleted session", zap.String("session_id", event.SessionDeleted.ID))
		err := m.store.Delete(event.SessionDeleted.ID)
		return err
	default:
		return errors.New("invalid event type received")
	}
}
