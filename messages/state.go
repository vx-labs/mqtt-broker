package queues

import (
	"errors"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/messages/pb"
	"go.uber.org/zap"
)

const (
	MessagePut string = "message_put"
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

func (m *server) applyEvent(event *pb.MessagesStateTransition) error {
	switch event.Kind {
	case MessagePut:
		err := errors.New("not implem")
		if err != nil {
			m.logger.Error("failed to put message in queue", zap.Error(err))
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
		m.logger.Error("failed to commit message ack event", zap.Error(err))
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
			m.logger.Error("failed to apply event", zap.String("event_kidn", event.Kind))
			return err
		}
	}
	return nil
}
