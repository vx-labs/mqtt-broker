package sessions

import (
	"bytes"
	"errors"
	fmt "fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/golang/protobuf/proto"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/sessions/pb"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

const (
	transitionSessionCreated = "session_created"
	transitionSessionDeleted = "session_deleted"
)

func (b *server) Shutdown() {
	err := b.state.Shutdown()
	if err != nil {
		b.logger.Error("failed to shutdown raft instance cleanly", zap.Error(err))
	}
	b.gprcServer.GracefulStop()
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	err := b.state.Start(name, b)
	if err != nil {
		panic(err)
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			b.gcExpiredSessions()
		}
	}()
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

func (m *server) Snapshot() io.ReadCloser {
	dump, err := m.store.All(nil)
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

	return ioutil.NopCloser(bytes.NewReader(payload))
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
	pb.RegisterSessionsServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	m.gprcServer = s
	return lis
}

func (m *server) Apply(payload []byte) error {
	event := pb.SessionStateTransition{}
	err := proto.Unmarshal(payload, &event)
	if err != nil {
		return err
	}
	switch event.Kind {
	case transitionSessionCreated:
		input := event.SessionCreated.Input
		err := m.store.Create(input)
		return err
	case transitionSessionDeleted:
		err := m.store.Delete(event.SessionDeleted.ID)
		return err
	default:
		return errors.New("invalid event type received")
	}
}
