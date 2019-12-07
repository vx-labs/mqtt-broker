package kv

import (
	"io"

	"errors"

	"github.com/gogo/protobuf/proto"

	"github.com/vx-labs/mqtt-broker/kv/pb"
	"go.uber.org/zap"
)

func (m *server) Restore(r io.Reader) error {
	err := m.store.Restore(r)
	if err != nil {
		m.logger.Error("failed to restore snapshot", zap.Error(err))
		return err
	}
	m.logger.Info("restored snapshot")
	return nil
}
func (m *server) Snapshot() io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		m.store.WriteTo(w)
	}()
	return r
}
func (m *server) applyEvent(payload *pb.KVStateTransition) error {
	switch event := payload.GetEvent().(type) {
	case *pb.KVStateTransition_DeleteBatch:
		input := event.DeleteBatch
		return m.store.DeleteBatch(input.KeyMDs)
	case *pb.KVStateTransition_Delete:
		input := event.Delete
		return m.store.Delete(input.Key, input.Version)
	case *pb.KVStateTransition_Set:
		input := event.Set
		return m.store.Put(input.Key, input.Value, input.Deadline, input.Version)
	default:
		return errors.New("invalid event received")
	}
}
func (m *server) commitEvent(payload ...*pb.KVStateTransition) error {
	event, err := proto.Marshal(&pb.KVStateTransitionSet{Events: payload})
	if err != nil {
		m.logger.Error("failed to encode event", zap.Error(err))
		return err
	}
	err = m.state.ApplyEvent(event)
	if err != nil {
		return err
	}
	return nil
}
func (m *server) Apply(payload []byte) error {
	data := pb.KVStateTransitionSet{}
	err := proto.Unmarshal(payload, &data)
	if err != nil {
		m.logger.Error("failed to unmarshal raft event", zap.Error(err))
		return err
	}
	for _, event := range data.Events {
		err := m.applyEvent(event)
		if err != nil {
			return err
		}
	}
	return nil
}
