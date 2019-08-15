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
	b.state = l
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

func (m *server) ByID(ctx context.Context, input *SessionByIDInput) (*Metadata, error) {
	session, err := m.store.ByID(input.ID)
	if err != nil {
		return nil, err
	}
	return &session.Metadata, nil
}
func (m *server) ByClientID(ctx context.Context, input *SessionByClientIDInput) (*SessionMetadataList, error) {
	sessions, err := m.store.ByClientID(input.ClientID)
	if err != nil {
		return nil, err
	}
	set := &SessionMetadataList{
		Metadatas: make([]*Metadata, len(sessions)),
	}
	for idx := range sessions {
		set.Metadatas[idx] = &sessions[idx].Metadata
	}
	return set, nil
}
func (m *server) ByPeer(ctx context.Context, input *SessionByPeerInput) (*SessionMetadataList, error) {
	sessions, err := m.store.ByPeer(input.Peer)
	if err != nil {
		return nil, err
	}
	set := &SessionMetadataList{
		Metadatas: make([]*Metadata, len(sessions)),
	}
	for idx := range sessions {
		set.Metadatas[idx] = &sessions[idx].Metadata
	}
	return set, nil
}
func (m *server) All(ctx context.Context, input *SessionFilterInput) (*SessionMetadataList, error) {
	sessions, err := m.store.All()
	if err != nil {
		return nil, err
	}
	set := &SessionMetadataList{
		Metadatas: make([]*Metadata, len(sessions)),
	}
	for idx := range sessions {
		set.Metadatas[idx] = &sessions[idx].Metadata
	}
	return set, nil
}
func (m *server) Create(ctx context.Context, input *SessionCreateInput) (*SessionCreateOutput, error) {
	ev := SessionStateTransition{
		Kind: transitionSessionCreated,
		SessionCreated: &SessionStateTransitionSessionCreated{
			Input: &SessionCreateInput{
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
		},
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return nil, err
	}
	return &SessionCreateOutput{}, m.state.ApplyEvent(payload)
}

func (m *server) Apply(payload []byte, leader bool) error {
	event := SessionStateTransition{}
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
