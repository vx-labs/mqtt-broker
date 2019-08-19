package sessions

import (
	"bytes"
	"errors"
	fmt "fmt"
	"io"
	"io/ioutil"
	"net"

	"github.com/golang/protobuf/proto"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/sessions/pb"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions/pb"

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
	subscriptionsConn, err := mesh.DialService("subscriptions")
	if err != nil {
		panic(err)
	}
	b.Subscriptions = subscriptions.NewClient(subscriptionsConn)
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	err = b.state.Start(name, b)
	if err != nil {
		panic(err)
	}
}
func (m *server) Restore(r io.Reader) error {
	payload, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	snapshot := &pb.SessionMetadataList{}
	err = proto.Unmarshal(payload, snapshot)
	if err != nil {
		return err
	}
	for _, session := range snapshot.Sessions {
		m.store.Create(session)
	}
	m.logger.Info("restored snapshot", zap.Int("size", len(payload)))
	return nil
}
func (m *server) Snapshot() io.Reader {
	dump, err := m.store.All()
	if err != nil {
		m.logger.Error("failed to snapshot store", zap.Error(err))
		return nil
	}
	payload, err := proto.Marshal(dump)
	if err != nil {
		m.logger.Error("failed to marshal snapshot", zap.Error(err))
		return nil
	}
	m.logger.Info("snapshotted store", zap.Int("size", len(payload)))
	return bytes.NewReader(payload)
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
		if leader {
			m.logger.Info("created session", zap.String("session_id", event.SessionCreated.Input.ID))
		}
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
		if leader {
			m.logger.Info("deleted session", zap.String("session_id", event.SessionDeleted.ID))
		}
		err := m.store.Delete(event.SessionDeleted.ID)
		if leader {
			set, err := m.Subscriptions.BySession(m.ctx, event.SessionDeleted.ID)
			if err != nil {
				m.logger.Error("failed to fetch session subscriptions", zap.String("session_id", event.SessionDeleted.ID), zap.Error(err))
			}
			for _, subscription := range set {
				err = m.Subscriptions.Delete(m.ctx, subscription.ID)
				if err != nil {
					m.logger.Error("failed to delete session subscription", zap.String("session_id", event.SessionDeleted.ID), zap.Error(err))
					break
				}
			}
		}
		return err
	default:
		return errors.New("invalid event type received")
	}
}
