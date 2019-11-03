package queues

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	"github.com/vx-labs/mqtt-broker/queues/store"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/queues/pb"
	"go.uber.org/zap"
)

const (
	QueueCreated    string = "queue_created"
	QueueDeleted    string = "queue_delete"
	QueueMessagePut string = "queue_message_put"
)

func (m *server) Restore(r io.Reader) error {
	payload, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	snapshot := &pb.QueueMetadataList{}
	err = proto.Unmarshal(payload, snapshot)
	if err != nil {
		return err
	}
	for _, session := range snapshot.Queues {
		m.store.CreateQueue(session.ID)
	}
	m.logger.Info("restored snapshot", zap.Int("size", len(payload)))
	return nil
}
func (m *server) Snapshot() io.Reader {
	queueList := m.store.All()
	state := &pb.QueueMetadataList{
		Queues: make([]*pb.QueueMetadata, len(queueList)),
	}
	for idx := range queueList {
		state.Queues[idx] = &pb.QueueMetadata{ID: queueList[idx]}
	}
	payload, err := proto.Marshal(state)
	if err != nil {
		m.logger.Error("failed to marshal snapshot", zap.Error(err))
		return nil
	}
	m.logger.Info("snapshotted store", zap.Int("size", len(payload)))
	return bytes.NewReader(payload)
}

func (m *server) Apply(payload []byte) error {
	event := pb.QueuesStateTransition{}
	err := proto.Unmarshal(payload, &event)
	if err != nil {
		return err
	}
	switch event.Kind {
	case QueueCreated:
		input := event.QueueCreated.Input
		err := m.store.CreateQueue(input.ID)
		return err
	case QueueDeleted:
		input := event.QueueDeleted
		m.store.DeleteQueue(input.ID)
		return nil
	case QueueMessagePut:
		input := event.QueueMessagePut
		err := m.store.Put(input.QueueID, input.Offset, input.Payload)
		if err == store.ErrQueueNotFound {
			return nil
		}
		return err
	default:
		return errors.New("invalid event type received")
	}
}
