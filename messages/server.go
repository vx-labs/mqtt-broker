package messages

import (
	"context"
	fmt "fmt"
	"hash/fnv"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"

	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/messages/pb"
	"github.com/vx-labs/mqtt-broker/messages/store"
	"go.uber.org/zap"
)

const (
	dbPath = "/var/tmp/mqtt-broker-messages"
)

type server struct {
	id         string
	store      *store.BoltStore
	state      types.RaftServiceLayer
	ctx        context.Context
	gprcServer *grpc.Server
	logger     *zap.Logger
	leaderRPC  pb.MessagesServiceClient
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

func (s *server) DeleteStream(ctx context.Context, input *pb.MessageDeleteStreamInput) (*pb.MessageDeleteStreamOutput, error) {
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.DeleteStream(ctx, input)
	}
	err := s.commitEvent(&pb.MessagesStateTransition{
		Event: &pb.MessagesStateTransition_StreamDeleted{
			StreamDeleted: &pb.MessagesStateTransitionStreamDeleted{
				StreamID: input.ID,
			},
		},
	})
	if err != nil {
		s.logger.Error("failed to commit stream deleted event", zap.Error(err))
	}
	return &pb.MessageDeleteStreamOutput{}, err
}
func (s *server) CreateStream(ctx context.Context, input *pb.MessageCreateStreamInput) (*pb.MessageCreateStreamOutput, error) {
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.CreateStream(ctx, input)
	}
	if input.ID == "" {
		return nil, ErrInvalidArgument("field 'ID' is required")
	}
	if input.ShardCount < 1 {
		return nil, ErrInvalidArgument("ShardCount must be at least 1")
	}
	config := pb.StreamConfig{
		ID:       input.ID,
		ShardIDs: make([]string, input.ShardCount),
	}
	for idx := range config.ShardIDs {
		config.ShardIDs[idx] = uuid.New().String()
	}
	err := s.commitEvent(&pb.MessagesStateTransition{
		Event: &pb.MessagesStateTransition_StreamCreated{
			StreamCreated: &pb.MessagesStateTransitionStreamCreated{
				Config: &config,
			},
		},
	})
	if err != nil {
		s.logger.Error("failed to commit stream created event", zap.Error(err))
	}
	return &pb.MessageCreateStreamOutput{}, err
}

func ErrNotFound(msg string) error {
	return status.Error(codes.NotFound, msg)
}

func ErrInvalidArgument(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

func (s *server) GetStream(ctx context.Context, input *pb.MessageGetStreamInput) (*pb.MessageGetStreamOutput, error) {
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	config := s.store.GetStream(input.ID)
	if config == nil {
		return nil, ErrNotFound("stream not found")
	}
	return &pb.MessageGetStreamOutput{ID: input.ID, Config: config}, nil
}
func (s *server) ListStreams(ctx context.Context, input *pb.MessageListStreamInput) (*pb.MessageListStreamOutput, error) {
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	streams := s.store.ListStreams()
	return &pb.MessageListStreamOutput{Streams: streams}, nil
}
func (s *server) GetMessages(ctx context.Context, input *pb.MessageGetMessagesInput) (*pb.MessageGetMessagesOutput, error) {
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	buf := make([]*pb.StoredMessage, input.MaxCount)
	count, offset, err := s.store.GetRange(input.StreamID, input.ShardID, input.Offset, buf)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		offset = input.Offset
	}
	return &pb.MessageGetMessagesOutput{
		Messages:   buf[:count],
		NextOffset: offset,
		StreamID:   input.StreamID,
		ShardID:    input.ShardID,
	}, nil
}

func (s *server) PutMessage(ctx context.Context, input *pb.MessagePutMessageInput) (*pb.MessagePutMessageOutput, error) {
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.PutMessage(ctx, input)
	}
	stream := s.store.GetStream(input.StreamID)
	if stream == nil {
		return nil, ErrNotFound("stream not found")
	}
	shardIdx := hashShardKey(input.ShardKey, len(stream.ShardIDs))
	shardID := stream.ShardIDs[shardIdx]
	// FIXME: do not use time as a key generator
	idx := uint64(time.Now().UnixNano())
	payload := input.Payload
	err := s.commitEvent(&pb.MessagesStateTransition{
		Event: &pb.MessagesStateTransition_MessagePut{
			MessagePut: &pb.MessagesStateTransitionMessagePut{
				StreamID: input.StreamID,
				ShardID:  shardID,
				Offset:   idx,
				Payload:  payload,
			},
		},
	})
	if err != nil {
		s.logger.Error("failed to commit queue put event", zap.Error(err))
	}
	return &pb.MessagePutMessageOutput{
		Offset:   idx,
		ShardID:  shardID,
		StreamID: stream.ID,
	}, err
}
func (s *server) PutMessageBatch(ctx context.Context, ev *pb.MessagePutMessageBatchInput) (*pb.MessagePutMessageBatchOutput, error) {
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	if !s.state.IsLeader() {
		return s.leaderRPC.PutMessageBatch(ctx, ev)
	}
	event := []*pb.MessagesStateTransition{}
	for _, input := range ev.Batches {
		stream := s.store.GetStream(input.StreamID)
		if stream == nil {
			return nil, ErrNotFound("stream not found")
		}
		idx := uint64(time.Now().UnixNano())
		payload := input.Payload
		shardIdx := hashShardKey(input.ShardKey, len(stream.ShardIDs))
		shardID := stream.ShardIDs[shardIdx]
		event = append(event, &pb.MessagesStateTransition{
			Event: &pb.MessagesStateTransition_MessagePut{
				MessagePut: &pb.MessagesStateTransitionMessagePut{
					StreamID: input.StreamID,
					ShardID:  shardID,
					Offset:   idx,
					Payload:  payload,
				},
			},
		})
		idx++
	}
	err := s.commitEvent(event...)
	if err != nil {
		s.logger.Error("failed to commit messages put batch event", zap.Error(err))
	}
	return &pb.MessagePutMessageBatchOutput{}, err
}
