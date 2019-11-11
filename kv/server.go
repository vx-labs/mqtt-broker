package kv

import (
	"context"
	fmt "fmt"
	"time"

	grpc "google.golang.org/grpc"
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
	id         string
	store      *store.BoltStore
	state      types.RaftServiceLayer
	ctx        context.Context
	gprcServer *grpc.Server
	logger     *zap.Logger
	leaderRPC  pb.KVServiceClient
}

func New(id string, logger *zap.Logger) *server {
	logger.Debug("opening kv durable store")
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
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.Delete(ctx, input)
	}
	md, err := s.store.GetMetadata(input.Key)
	if err != nil {
		return nil, err
	}
	if md.Version != input.Version {
		return nil, status.Error(codes.FailedPrecondition, "version not matched")
	}

	err = s.commitEvent(&pb.KVStateTransition{
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
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.Set(ctx, input)
	}
	if len(input.Key) == 0 {
		return nil, ErrInvalidArgument("field 'Key' is required")
	}
	if len(input.Value) == 0 {
		return nil, ErrInvalidArgument("field 'Value' is required")
	}
	md, err := s.store.GetMetadata(input.Key)
	if err != nil {
		return nil, err
	}
	if md.Version != input.Version {
		return nil, status.Error(codes.FailedPrecondition, "version mismatched")
	}
	var deadline uint64 = 0
	if input.TimeToLive > 0 {
		deadline = uint64(time.Now().Add(time.Duration(input.TimeToLive)).UnixNano())
	}
	err = s.commitEvent(&pb.KVStateTransition{
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
		if err == store.ErrIndexOutdated {
			return nil, status.Error(codes.FailedPrecondition, "version mismatched")
		}
		s.logger.Warn("failed to commit value set event", zap.Error(err))
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
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.Get(ctx, input)
	}
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
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.GetMetadata(ctx, input)
	}
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
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.GetWithMetadata(ctx, input)
	}
	value, md, err := s.store.GetWithMetadata(input.Key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return nil, ErrNotFound("key not found")
		}
		return nil, err
	}
	return &pb.KVGetWithMetadataOutput{Value: value, Metadata: md}, nil
}
