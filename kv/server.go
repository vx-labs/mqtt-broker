package kv

import (
	"context"
	fmt "fmt"
	"hash/fnv"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/kv/pb"
	"github.com/vx-labs/mqtt-broker/kv/store"
	"go.uber.org/zap"
)

const (
	dbPath = "/var/tmp/mqtt-broker-kv"
)

type server struct {
	id        string
	store     *store.BoltStore
	state     types.RaftServiceLayer
	ctx       context.Context
	listeners []net.Listener
	logger    *zap.Logger
}

func hashShardKey(key string, shardCount int) int {
	hash := fnv.New32()
	hash.Write([]byte(key))
	return int(hash.Sum32()) % shardCount
}

func New(id string, logger *zap.Logger) *server {
	logger.Debug("opening messages durable store")
	boltstore, err := store.New(store.Options{
		NoSync: false,
		Path:   fmt.Sprintf("%s-%s-queues.bolt", dbPath, id),
	})
	if err != nil {
		logger.Fatal("failed to open messages durable store", zap.Error(err))
	}
	b := &server{
		id:     id,
		ctx:    context.Background(),
		store:  boltstore,
		logger: logger,
	}
	return b
}

func (s *server) Delete(ctx context.Context, input *pb.KVDeleteInput) (*pb.KVDeleteOutput, error) {
	err := s.commitEvent(&pb.KVStateTransition{
		Event: &pb.KVStateTransition_Delete{
			Delete: &pb.KVStateTransitionValueDeleted{
				Key: input.Key,
			},
		},
	})
	if err != nil {
		s.logger.Error("failed to commit value deleted event", zap.Error(err))
	}
	return &pb.KVDeleteOutput{}, err
}
func (s *server) Set(ctx context.Context, input *pb.KVSetInput) (*pb.KVSetOutput, error) {
	if len(input.Key) == 0 {
		return nil, ErrInvalidArgument("field 'Key' is required")
	}
	if len(input.Value) == 0 {
		return nil, ErrInvalidArgument("field 'Value' is required")
	}
	err := s.commitEvent(&pb.KVStateTransition{
		Event: &pb.KVStateTransition_Set{
			Set: &pb.KVStateTransitionValueSet{
				Key:   input.Key,
				Value: input.Value,
			},
		},
	})
	if err != nil {
		s.logger.Error("failed to commit value set event", zap.Error(err))
	}
	return &pb.KVSetOutput{}, err
}

func ErrNotFound(msg string) error {
	return status.Error(codes.NotFound, msg)
}

func ErrInvalidArgument(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

func (s *server) Get(ctx context.Context, input *pb.KVGetInput) (*pb.KVGetOutput, error) {
	value, err := s.store.Get(input.Key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, ErrNotFound("key not found")
	}
	return &pb.KVGetOutput{Key: input.Key, Value: value}, nil
}
