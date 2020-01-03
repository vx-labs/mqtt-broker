package queues

import (
	"context"
	fmt "fmt"
	"io"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/adapters/cp"
	"github.com/vx-labs/mqtt-broker/path"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/services/queues/pb"
	"github.com/vx-labs/mqtt-broker/services/queues/store"
	sessions "github.com/vx-labs/mqtt-broker/services/sessions/pb"
	"github.com/vx-labs/mqtt-broker/stream"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type SessionStore interface {
	ByID(ctx context.Context, id string) (*sessions.Session, error)
	All(ctx context.Context, filters ...sessions.SessionFilter) ([]*sessions.Session, error)
}

type server struct {
	id         string
	store      *store.BoltStore
	state      cp.Synchronizer
	ctx        context.Context
	gprcServer *grpc.Server
	logger     *zap.Logger
	sessions   SessionStore
	Messages   *messages.Client
	stream     *stream.Client
}

func New(id string, logger *zap.Logger) *server {
	logger.Debug("opening queues durable store")
	boltstore, err := store.New(store.Options{
		NoSync: false,
		Path:   fmt.Sprintf("%s/db.bolt", path.ServiceDataDir(id, "queues")),
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
	return s.commitEvent(&pb.QueuesStateTransition{
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
	return s.commitEvent(&pb.QueuesStateTransition{
		Kind: QueueDeleted,
		QueueDeleted: &pb.QueueStateTransitionQueueDeleted{
			ID: id,
		},
	})
}
func (s *server) clusterPutMessage(id string, payload []byte) error {
	if !s.store.Exists(id) {
		return store.ErrQueueNotFound
	}
	idx := time.Now().UnixNano()

	return s.commitEvent(&pb.QueuesStateTransition{
		Kind: QueueMessagePut,
		QueueMessagePut: &pb.QueueStateTransitionMessagePut{
			Offset:  uint64(idx),
			Payload: payload,
			QueueID: id,
		},
	})
}
func (s *server) clusterPutMessageBatch(ids []string, payloads [][]byte) error {
	event := []*pb.QueuesStateTransition{}
	for idx := range ids {
		offset := time.Now().UnixNano()
		if !s.store.Exists(ids[idx]) {
			return store.ErrQueueNotFound
		}
		payload := payloads[idx]
		event = append(event, &pb.QueuesStateTransition{
			Kind: QueueMessagePut,
			QueueMessagePut: &pb.QueueStateTransitionMessagePut{
				Offset:  uint64(offset),
				Payload: payload,
				QueueID: ids[idx],
			},
		})
		idx++
	}
	return s.commitEvent(event...)
}
func (s *server) PutMessage(ctx context.Context, input *pb.QueuePutMessageInput) (*pb.QueuePutMessageOutput, error) {
	payload, err := proto.Marshal(input.Publish)
	if err != nil {
		s.logger.Error("failed to encode message", zap.Error(err))
		return nil, err
	}
	err = s.clusterPutMessage(input.Id, payload)
	if err != nil {
		s.logger.Error("failed to commit queue put event", zap.Error(err))
	}
	return &pb.QueuePutMessageOutput{}, err
}
func (s *server) PutMessageBatch(ctx context.Context, ev *pb.QueuePutMessageBatchInput) (*pb.QueuePutMessageBatchOutput, error) {
	payloads := make([][]byte, len(ev.Batches))
	ids := make([]string, len(ev.Batches))
	for idx := range ev.Batches {
		payload, err := proto.Marshal(ev.Batches[idx].Publish)
		if err != nil {
			s.logger.Error("failed to encode message", zap.Error(err))
			return nil, err
		}
		payloads[idx] = payload
		ids[idx] = ev.Batches[idx].Id
	}
	err := s.clusterPutMessageBatch(ids, payloads)
	if err != nil {
		s.logger.Error("failed to commit queue put batch event", zap.Error(err))
	}
	return &pb.QueuePutMessageBatchOutput{}, err
}
func (s *server) StreamMessages(input *pb.QueueGetMessagesInput, stream pb.QueuesService_StreamMessagesServer) error {
	offset := input.Offset
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
	batchSize := 10
	items := make([]store.StoredMessage, batchSize)
	count := 0
	for {
		count, offset, err = s.store.GetRange(input.Id, 0, items)
		if err != nil {
			return err
		}
		out := make([]*pb.QueueGetMessagesOutput, count)
		events := make([]*pb.QueuesStateTransition, count)
		for idx := range out {
			inflightDeadline := uint64(time.Now().Add(30 * time.Second).UnixNano())
			out[idx] = &pb.QueueGetMessagesOutput{
				Id:        input.Id,
				Offset:    offset,
				AckOffset: inflightDeadline,
				Publish:   &packet.Publish{},
			}
			err = proto.Unmarshal(items[idx].Payload, out[idx].Publish)
			if err != nil {
				return err
			}
			events[idx] = &pb.QueuesStateTransition{
				Kind: QueueMessageSetInflight,
				MessageInflightSet: &pb.QueueStateTransitionMessageInflightSet{
					Deadline: inflightDeadline,
					Offset:   items[idx].Offset,
					QueueID:  input.Id,
				},
			}
		}
		err := s.commitEvent(events...)
		if err != nil {
			s.logger.Error("failed to commit message ack event", zap.Error(err))
			return err
		}
		err = stream.Send(&pb.StreamMessageOutput{
			Offset:  offset,
			Batches: out,
		})
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if count == batchSize {
			continue
		}
		select {
		case <-tick:
		case <-closed:
			return nil
		}
	}
}
func (s *server) AckMessage(ctx context.Context, input *pb.AckMessageInput) (*pb.AckMessageOutput, error) {
	err := s.commitEvent(&pb.QueuesStateTransition{
		Kind: QueueMessageAcked,
		MessageAcked: &pb.QueueStateTransitionMessageAcked{
			Offset:  input.Offset,
			QueueID: input.Id,
		},
	})
	if err != nil {
		s.logger.Error("failed to commit message ack event", zap.Error(err))
	}
	return &pb.AckMessageOutput{}, err
}
func (s *server) List(ctx context.Context, input *pb.QueuesListInput) (*pb.QueuesListOutput, error) {
	queues, err := s.store.ListQueues()
	if err != nil {
		return nil, err
	}
	return &pb.QueuesListOutput{QueueIDs: queues}, nil
}
func (s *server) GetStatistics(ctx context.Context, input *pb.QueueGetStatisticsInput) (*pb.QueueGetStatisticsOutput, error) {
	stats, err := s.store.GetStatistics(input.ID)
	if err != nil {
		return nil, err
	}
	return &pb.QueueGetStatisticsOutput{Statistics: stats}, nil
}
func contains(needle string, slice []string) bool {
	for _, s := range slice {
		if s == needle {
			return true
		}
	}
	return false
}
