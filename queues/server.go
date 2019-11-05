package queues

import (
	"context"
	fmt "fmt"
	"io"
	"net"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/queues/pb"
	"github.com/vx-labs/mqtt-broker/queues/store"
	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

const (
	dbPath = "/var/tmp/mqtt-broker-queues"
)

type SessionStore interface {
	ByID(ctx context.Context, id string) (*sessions.Session, error)
}

type server struct {
	id        string
	store     *store.BoltStore
	state     types.RaftServiceLayer
	ctx       context.Context
	listeners []net.Listener
	logger    *zap.Logger
	sessions  SessionStore
}

func New(id string, logger *zap.Logger) *server {
	logger.Debug("opening queues durable store")
	boltstore, err := store.New(store.Options{
		NoSync: false,
		Path:   fmt.Sprintf("%s-%s-queues.bolt", dbPath, id),
	})
	if err != nil {
		logger.Fatal("failed to open queues durable store", zap.Error(err))
	}
	b := &server{
		id:     id,
		ctx:    context.Background(),
		store:  boltstore,
		logger: logger,
	}
	return b
}

func (s *server) Create(ctx context.Context, input *pb.QueueCreateInput) (*pb.QueueCreateOutput, error) {
	return &pb.QueueCreateOutput{}, s.clusterCreateQueue(input.Id)
}
func (s *server) Delete(ctx context.Context, input *pb.QueueDeleteInput) (*pb.QueueDeleteOutput, error) {
	return &pb.QueueDeleteOutput{}, s.clusterDeleteQueue(input.Id)
}

func (s *server) clusterCreateQueue(id string) error {
	if s.store.Exists(id) {
		return nil
	}
	return s.applyTransition(&pb.QueuesStateTransition{
		Kind: QueueCreated,
		QueueCreated: &pb.QueueStateTransitionQueueCreated{
			Input: &pb.QueueMetadata{ID: id},
		},
	})
}
func (s *server) clusterDeleteQueue(id string) error {
	if !s.store.Exists(id) {
		return nil
	}
	return s.applyTransition(&pb.QueuesStateTransition{
		Kind: QueueDeleted,
		QueueDeleted: &pb.QueueStateTransitionQueueDeleted{
			ID: id,
		},
	})
}
func (s *server) applyTransition(event *pb.QueuesStateTransition) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		s.logger.Error("failed to encode event", zap.Error(err), zap.String("event_kind", event.Kind))
		return err
	}
	err = s.state.ApplyEvent(payload)
	if err != nil {
		s.logger.Error("failed to commit event", zap.Error(err), zap.String("event_kind", event.Kind))
	}
	return err
}
func (s *server) PutMessage(ctx context.Context, input *pb.QueuePutMessageInput) (*pb.QueuePutMessageOutput, error) {
	// FIXME: do not use time as a key generator
	idx := time.Now().UnixNano()
	payload, err := proto.Marshal(input.Publish)
	if err != nil {
		s.logger.Error("failed to encode message", zap.Error(err))
		return nil, err
	}
	event, err := proto.Marshal(&pb.QueuesStateTransition{
		Kind: QueueMessagePut,
		QueueMessagePut: &pb.QueueStateTransitionMessagePut{
			Offset:  uint64(idx),
			Payload: payload,
			QueueID: input.Id,
		},
	})
	if err != nil {
		s.logger.Error("failed to encode event", zap.Error(err))
		return nil, err
	}
	err = s.state.ApplyEvent(event)
	if err != nil {
		s.logger.Error("failed to commit queue put event", zap.Error(err))
	}
	return &pb.QueuePutMessageOutput{}, err
}
func (s *server) PutMessageBatch(ctx context.Context, ev *pb.QueuePutMessageBatchInput) (*pb.QueuePutMessageBatchOutput, error) {
	// FIXME: do not use time as a key generator
	now := time.Now().UnixNano()
	event := &pb.QueuesStateTransition{
		Kind: QueueMessagePutBatch,
		QueueMessagePutBatch: &pb.QueueStateTransitionMessagePutBatch{
			Batches: []*pb.QueueStateTransitionMessagePut{},
		},
	}
	for _, input := range ev.Batches {
		idx := now
		payload, err := proto.Marshal(input.Publish)
		if err != nil {
			s.logger.Error("failed to encode message", zap.Error(err))
			return nil, err
		}
		event.QueueMessagePutBatch.Batches = append(event.QueueMessagePutBatch.Batches, &pb.QueueStateTransitionMessagePut{
			Offset:  uint64(idx),
			Payload: payload,
			QueueID: input.Id,
		})
	}
	payload, err := proto.Marshal(event)
	if err != nil {
		s.logger.Error("failed to encode event", zap.Error(err))
		return nil, err
	}
	err = s.state.ApplyEvent(payload)
	if err != nil {
		s.logger.Error("failed to commit queue put batch event", zap.Error(err))
	}
	return &pb.QueuePutMessageBatchOutput{}, err
}
func (s *server) StreamMessages(input *pb.QueueGetMessagesInput, stream pb.QueuesService_StreamMessagesServer) error {
	offset := input.Offset
	buff := make([][]byte, 10)
	count := 0
	var err error
	tick := make(chan struct{}, 1)
	closed := make(chan struct{})
	defer close(tick)

	cancelTicker := s.store.On(input.Id, "message_put", func(_ interface{}) {
		select {
		case tick <- struct{}{}:
		default:
		}
	})
	defer cancelTicker()
	queueDeletedTicker := s.store.On(input.Id, "queue_deleted", func(_ interface{}) {
		close(closed)
	})
	defer queueDeletedTicker()
	for {
		offset, count, err = s.store.GetRange(input.Id, offset, buff)
		if err != nil {
			return err
		}
		out := make([]*packet.Publish, count)
		for idx := range out {
			out[idx] = &packet.Publish{}
			err = proto.Unmarshal(buff[idx], out[idx])
			if err != nil {
				return err
			}
		}
		err = stream.Send(&pb.QueueGetMessagesOutput{
			Id:        input.Id,
			Offset:    offset,
			Publishes: out,
		})
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		select {
		case <-tick:
		case <-closed:
			return nil
		}
	}
}
func (s *server) GetMessages(ctx context.Context, input *pb.QueueGetMessagesInput) (*pb.QueueGetMessagesOutput, error) {
	buff := make([][]byte, 10)
	offset, count, err := s.store.GetRange(input.Id, input.Offset, buff)
	if err != nil {
		return nil, err
	}
	out := make([]*packet.Publish, count)
	for idx := range out {
		out[idx] = &packet.Publish{}
		err = proto.Unmarshal(buff[idx], out[idx])
		if err != nil {
			return nil, err
		}
	}
	return &pb.QueueGetMessagesOutput{
		Id:        input.Id,
		Offset:    offset,
		Publishes: out,
	}, nil
}
func (s *server) GetMessagesBatch(ctx context.Context, batchInput *pb.QueueGetMessagesBatchInput) (*pb.QueueGetMessagesBatchOutput, error) {
	in := make([]store.BatchInput, len(batchInput.Batches))
	for idx := range in {
		buff := make([][]byte, 10)
		in[idx].Data = buff
		in[idx].ID = batchInput.Batches[idx].Id
		in[idx].Offset = batchInput.Batches[idx].Offset
	}
	batchOut, err := s.store.GetRangeBatch(in)
	if err != nil {
		return nil, err
	}
	resp := &pb.QueueGetMessagesBatchOutput{}

	for idx, input := range batchOut {
		if input.Err != nil {
			continue
		}
		out := make([]*packet.Publish, input.Count)
		for innerIdx := range out {
			out[innerIdx] = &packet.Publish{}
			err = proto.Unmarshal(in[idx].Data[innerIdx], out[innerIdx])
			if err != nil {
				continue
			}
		}
		resp.Batches = append(resp.Batches, &pb.QueueGetMessagesOutput{
			Id:        input.ID,
			Offset:    input.Offset,
			Publishes: out,
		})
	}
	return resp, nil
}

func (m *server) isQueueExpired(id string) bool {
	_, err := m.sessions.ByID(m.ctx, id)
	return err != nil
}

func (m *server) gcExpiredQueues() {
	if m.Health() != "ok" {
		return
	}
	queues := m.store.All()
	for _, queue := range queues {
		if m.isQueueExpired(queue) {
			err := m.clusterDeleteQueue(queue)
			if err != nil {
				m.logger.Error("failed to gc queue", zap.String("queue_id", queue), zap.Error(err))
			} else {
				m.logger.Info("deleted expired queue", zap.String("queue_id", queue))
			}
		}
	}
}
