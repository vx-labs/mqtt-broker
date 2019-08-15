package sessions

import (
	"context"
	"errors"
	fmt "fmt"
	"io"
	"net"

	"github.com/golang/protobuf/proto"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"

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
	l := cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	err := l.Start(name, b)
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
	RegisterSessionsServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	return lis
}

func (m *server) Create(ctx context.Context, input *SessionCreateInput) (*SessionCreateOutput, error) {
	return &SessionCreateOutput{}, m.store.Upsert(Session{
		Metadata: Metadata{
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
		},
	}, nil)
}

func (m *server) Apply(payload []byte, leader bool) error {
	event := StateTransition{}
	err := proto.Unmarshal(payload, &event)
	if err != nil {
		return err
	}
	switch event.Kind {
	case transitionSessionCreated:
		input := event.SessionCreated.Input
		return m.store.Upsert(Session{
			Metadata: Metadata{
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
			},
		}, nil)
	case transitionSessionDeleted:
		return m.store.Delete(event.SessionDeleted.ID, "session deleted")

	default:
		return errors.New("invalid event type received")
	}
}
