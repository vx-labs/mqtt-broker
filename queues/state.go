package queues

import (
	"errors"
	"io"

	"github.com/vx-labs/mqtt-broker/queues/store"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/queues/pb"
	"go.uber.org/zap"
)

const (
	QueueCreated            string = "queue_created"
	QueueDeleted            string = "queue_delete"
	QueueMessagePut         string = "queue_message_put"
	QueueMessagePutBatch    string = "queue_message_put_batch"
	QueueMessageDeleted     string = "queue_message_deleted"
	QueueMessageSetInflight string = "queue_message_set_inflight"
	QueueMessageAcked       string = "queue_message_acked"
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

func (m *server) applyEvent(event *pb.QueuesStateTransition) error {
	switch event.Kind {
	case QueueCreated:
		input := event.QueueCreated.Input
		err := m.store.CreateQueue(input.ID)
		if err != nil {
			m.logger.Error("failed to create queue in store", zap.Error(err))
		}
		return err
	case QueueDeleted:
		input := event.QueueDeleted
		err := m.store.DeleteQueue(input.ID)
		if err != nil {
			m.logger.Error("failed to delete queue in store", zap.Error(err))
		}
		return nil
	case QueueMessagePut:
		input := event.QueueMessagePut
		err := m.store.Put(input.QueueID, input.Offset, input.Payload)
		if err == store.ErrQueueNotFound {
			return nil
		}
		if err != nil {
			m.logger.Error("failed to put message in queue", zap.Error(err))
		}
		return err
	case QueueMessagePutBatch:
		for _, input := range event.QueueMessagePutBatch.Batches {
			err := m.store.Put(input.QueueID, input.Offset, input.Payload)
			if err != nil {
				m.logger.Error("failed to put message in queue", zap.Error(err))
			}
		}
		return nil
	case QueueMessageAcked:
		input := event.MessageAcked
		err := m.store.AckInflight(input.QueueID, input.Offset)
		if err != nil {
			m.logger.Error("failed to ack message inflight", zap.Error(err))
		}
		return nil
	case QueueMessageSetInflight:
		input := event.MessageInflightSet
		err := m.store.SetInflight(input.QueueID, input.Offset, input.Deadline)
		if err != nil {
			m.logger.Error("failed to set message inflight", zap.Error(err))
		}
		return nil
	case QueueMessageDeleted:
		input := event.MessageDeleted
		err := m.store.Delete(input.QueueID, input.Offset)
		if err != nil {
			m.logger.Error("failed to delete message in queue", zap.Error(err))
		}
		return nil
	default:
		return errors.New("invalid event type received")
	}
}
func (m *server) commitEvent(payload ...*pb.QueuesStateTransition) error {
	event, err := proto.Marshal(&pb.QueuesStateTransitionSet{Events: payload})
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
	data := pb.QueuesStateTransitionSet{}
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
