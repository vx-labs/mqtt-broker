package messages

import (
	"errors"
	"io"

	"github.com/gogo/protobuf/proto"

	"github.com/vx-labs/mqtt-broker/services/messages/pb"
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
func (m *server) applyEvent(payload *pb.MessagesStateTransition) error {
	switch event := payload.GetEvent().(type) {
	case *pb.MessagesStateTransition_StreamDeleted:
		input := event.StreamDeleted
		err := m.store.DeleteStream(input.StreamID)
		if err != nil {
			m.logger.Error("failed to delete stream", zap.Error(err))
		}
		return err
	case *pb.MessagesStateTransition_StreamCreated:
		input := event.StreamCreated
		err := m.store.CreateStream(input.Config)
		if err != nil {
			m.logger.Error("failed to create stream", zap.Error(err))
		}
		return err
	case *pb.MessagesStateTransition_MessagePut:
		input := event.MessagePut
		err := m.store.Put(input.StreamID, input.ShardID, input.Offset, input.Payload)
		if err != nil {
			m.logger.Error("failed to store message", zap.Error(err))
		}
		return err
	default:
		return errors.New("invalid event type received")
	}
}
func (m *server) commitEvent(payload ...*pb.MessagesStateTransition) error {
	event, err := proto.Marshal(&pb.MessagesStateTransitionSet{Events: payload})
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
	data := pb.MessagesStateTransitionSet{}
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
