package subscriptions

import (
	"bytes"
	"errors"
	fmt "fmt"
	"io"
	"io/ioutil"
	"net"

	"github.com/golang/protobuf/proto"
	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/subscriptions/pb"

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
	err := b.state.Shutdown()
	if err != nil {
		b.logger.Error("failed to shutdown raft instance cleanly", zap.Error(err))
	}
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	sessionsConn, err := mesh.DialService("sessions")
	if err != nil {
		panic(err)
	}
	b.sessions = sessions.NewClient(sessionsConn)
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	go func() {
		err := b.state.Start(name, b)
		if err != nil {
			panic(err)
		}
	}()
}
func (m *server) Restore(r io.Reader) error {
	payload, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	snapshot := &pb.SubscriptionMetadataList{}
	err = proto.Unmarshal(payload, snapshot)
	if err != nil {
		return err
	}
	for _, subscription := range snapshot.Metadatas {
		m.store.Create(subscription)
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
	return lis
}

func (m *server) Apply(payload []byte, leader bool) error {
	event := pb.SubscriptionStateTransition{}
	err := proto.Unmarshal(payload, &event)
	if err != nil {
		return err
	}
	switch event.Kind {
	case transitionSessionCreated:
		if leader {
			m.logger.Info("created subscription", zap.String("subscription_id", event.SubscriptionCreated.Input.ID))
		}
		input := event.SubscriptionCreated.Input
		err := m.store.Create(&pb.Metadata{
			ID:        input.ID,
			Peer:      input.Peer,
			Tenant:    input.Tenant,
			Pattern:   input.Pattern,
			SessionID: input.SessionID,
			Qos:       input.Qos,
		})
		return err
	case transitionSessionDeleted:
		if leader {
			m.logger.Info("deleted subscription", zap.String("subscription_id", event.SubscriptionDeleted.ID))
		}
		err := m.store.Delete(event.SubscriptionDeleted.ID)
		return err
	default:
		return errors.New("invalid event type received")
	}
}
