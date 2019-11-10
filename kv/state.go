package kv

import (
	"errors"
	"io"

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
func (m *server) Snapshot() io.Reader {
	r, w := io.Pipe()
	go m.store.WriteTo(w)
	return r
}
func (m *server) applyEvent(payload *pb.KVStateTransition) error {
	switch event := payload.GetEvent().(type) {
	case *pb.KVStateTransition_DeleteBatch:
		input := event.DeleteBatch
		err := m.store.DeleteBatch(input.Keys)
		if err != nil {
			m.logger.Error("failed to delete keys", zap.Error(err))
		}
		return err
	case *pb.KVStateTransition_Delete:
		input := event.Delete
		err := m.store.Delete(input.Key)
		if err != nil {
			m.logger.Error("failed to delete key", zap.Error(err))
		}
		return err
	case *pb.KVStateTransition_Set:
		input := event.Set
		err := m.store.Put(input.Key, input.Value, input.Deadline)
		if err != nil {
			m.logger.Error("failed to set key", zap.Error(err))
		}
		return err
	default:
		return errors.New("invalid event type received")
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
		m.logger.Error("failed to commit event", zap.Error(err))
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
			m.logger.Error("failed to apply event", zap.Error(err))
			return err
		}
	}
	return nil
}
