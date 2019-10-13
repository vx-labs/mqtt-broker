package queues

import (
	"context"
	"net"
	"time"

	"github.com/gogo/protobuf/proto"
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
	ctx       context.Context
	listeners []net.Listener
	logger    *zap.Logger
	sessions  SessionStore
}

func New(id string, logger *zap.Logger) *server {
	boltstore, err := store.New(store.Options{
		NoSync: false,
		Path:   dbPath,
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
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			b.gcExpiredQueues()
		}
	}()
	return b
}

func (s *server) Create(ctx context.Context, input *pb.QueueCreateInput) (*pb.QueueCreateOutput, error) {
	err := s.store.CreateQueue(input.Id)
	if err != nil {
		s.logger.Error("failed to create queue", zap.Error(err))
	} else {
		s.logger.Info("queue created", zap.String("queue_id", input.Id))
	}
	return &pb.QueueCreateOutput{}, err
}
func (s *server) Delete(ctx context.Context, input *pb.QueueDeleteInput) (*pb.QueueDeleteOutput, error) {
	err := s.store.DeleteQueue(input.Id)
	if err != nil {
		s.logger.Error("failed to delete queue", zap.Error(err))
	} else {
		s.logger.Info("queue deleted", zap.String("queue_id", input.Id))
	}
	return &pb.QueueDeleteOutput{}, err
}
func (s *server) PutMessage(ctx context.Context, input *pb.QueuePutMessageInput) (*pb.QueuePutMessageOutput, error) {
	// FIXME: do not use time as a key generator
	idx := time.Now().UnixNano()
	payload, err := proto.Marshal(input.Publish)
	if err != nil {
		s.logger.Error("failed to encode message", zap.Error(err))
		return nil, err
	}
	err = s.store.Put(input.Id, uint64(idx), payload)
	if err != nil {
		s.logger.Error("failed to enqueue message", zap.Error(err))
	}
	return &pb.QueuePutMessageOutput{}, err
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

	queues := m.store.All()
	for _, queue := range queues {
		if m.isQueueExpired(queue) {
			err := m.store.DeleteQueue(queue)
			if err != nil {
				m.logger.Error("failed to gc queue", zap.String("queue_id", queue), zap.Error(err))
			} else {
				m.logger.Info("deleted expired queue", zap.String("queue_id", queue))
			}
		}
	}
}
