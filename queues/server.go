package queues

import (
	"context"
	"net"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/queues/pb"
	"github.com/vx-labs/mqtt-broker/queues/store"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

const (
	dbPath = "/var/tmp/mqtt-broker-queues"
)

type server struct {
	id        string
	store     *store.BoltStore
	ctx       context.Context
	listeners []net.Listener
	logger    *zap.Logger
}

func New(id string, logger *zap.Logger) *server {
	boltstore, err := store.New(store.Options{
		NoSync: false,
		Path:   dbPath,
	})
	if err != nil {
		logger.Fatal("failed to open queues durable store", zap.Error(err))
	}
	return &server{
		id:     id,
		ctx:    context.Background(),
		store:  boltstore,
		logger: logger,
	}
}

func (s *server) Create(ctx context.Context, input *pb.QueueCreateInput) (*pb.QueueCreateOutput, error) {
	err := s.store.CreateQueue(input.Id)
	return &pb.QueueCreateOutput{}, err
}
func (s *server) Delete(ctx context.Context, input *pb.QueueDeleteInput) (*pb.QueueDeleteOutput, error) {
	err := s.store.DeleteQueue(input.Id)
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
		Offset:    offset,
		Publishes: out,
	}, nil
}
