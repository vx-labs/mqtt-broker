package kv

import (
	"context"
	fmt "fmt"
	"hash/fnv"
	"net"
	"time"

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
	if input.Version > 0 {
		md, err := s.store.GetMetadata(input.Key)
		if err != nil {
			return nil, err
		}
		if md.Version != input.Version {
			return nil, status.Error(codes.FailedPrecondition, "version not matched")
		}
	}
	err := s.commitEvent(&pb.KVStateTransition{
		Event: &pb.KVStateTransition_Delete{
			Delete: &pb.KVStateTransitionValueDeleted{
				Key:     input.Key,
				Version: input.Version,
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
	if input.Version > 0 {
		md, err := s.store.GetMetadata(input.Key)
		if err != nil {
			return nil, err
		}
		if md.Version != input.Version {
			return nil, status.Error(codes.FailedPrecondition, "version not matched")
		}
	}
	var deadline uint64 = 0
	if input.TimeToLive > 0 {
		deadline = uint64(time.Now().Add(time.Duration(input.TimeToLive)).UnixNano())
	}
	err := s.commitEvent(&pb.KVStateTransition{
		Event: &pb.KVStateTransition_Set{
			Set: &pb.KVStateTransitionValueSet{
				Key:      input.Key,
				Value:    input.Value,
				Deadline: deadline,
				Version:  input.Version,
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
func (s *server) GetMetadata(ctx context.Context, input *pb.KVGetMetadataInput) (*pb.KVGetMetadataOutput, error) {
	value, err := s.store.GetMetadata(input.Key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return nil, ErrNotFound("key not found")
		}
		return nil, err
	}
	return &pb.KVGetMetadataOutput{Metadata: value}, nil
}
func (s *server) GetWithMetadata(ctx context.Context, input *pb.KVGetWithMetadataInput) (*pb.KVGetWithMetadataOutput, error) {
	value, md, err := s.store.GetWithMetadata(input.Key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return nil, ErrNotFound("key not found")
		}
		return nil, err
	}
	return &pb.KVGetWithMetadataOutput{Value: value, Metadata: md}, nil
}
