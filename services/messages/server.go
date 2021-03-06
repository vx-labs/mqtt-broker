package messages

import (
	"context"
	fmt "fmt"
	"hash/fnv"
	"net"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/asaskevich/EventBus"
	"github.com/google/uuid"

	"github.com/vx-labs/mqtt-broker/adapters/cp"
	"github.com/vx-labs/mqtt-broker/path"
	"github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/services/messages/store"
	"go.uber.org/zap"
)

type messagePutNotification struct {
	StreamID string
	ShardID  string
	Offset   uint64
}
type server struct {
	id           string
	store        *store.BoltStore
	eventBus     EventBus.Bus
	messagePutCh chan *pb.MessagesStateTransitionMessagePut
	state        cp.Synchronizer
	ctx          context.Context
	gprcServer   *grpc.Server
	listener     net.Listener
	logger       *zap.Logger
	config       ServerConfig
	leaderRPC    pb.MessagesServiceClient
}

type ServerStreamConfig struct {
	ID         string
	ShardCount int64
}
type ServerConfig struct {
	InitialsStreams []ServerStreamConfig
}

func hashShardKey(key string, shardCount int) int {
	hash := fnv.New32()
	hash.Write([]byte(key))
	return int(hash.Sum32()) % shardCount
}
func New(id string, config ServerConfig, logger *zap.Logger) *server {
	logger.Debug("opening messages durable store")
	boltstore, err := store.New(store.Options{
		NoSync: false,
		Path:   fmt.Sprintf("%s/db.bolt", path.ServiceDataDir(id, "messages")),
	})
	if err != nil {
		logger.Fatal("failed to open messages durable store", zap.Error(err))
	}
	b := &server{
		id:           id,
		ctx:          context.Background(),
		eventBus:     EventBus.New(),
		messagePutCh: make(chan *pb.MessagesStateTransitionMessagePut, 10),
		store:        boltstore,
		logger:       logger,
		config:       config,
	}
	go func() {
		for notification := range b.messagePutCh {
			b.eventBus.Publish(fmt.Sprintf("%s/message_put", notification.ShardID), notification)
		}
	}()
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
func (s *server) GetStreamStatistics(ctx context.Context, input *pb.MessageGetStreamStatisticsInput) (*pb.MessageGetStreamStatisticsOutput, error) {
	if s.state == nil || s.leaderRPC == nil {
		return nil, status.Error(codes.Unavailable, "node is not ready")
	}
	stats, err := s.store.GetStreamStatistics(input.ID)
	if err == store.ErrStreamNotFound {
		return nil, ErrNotFound("stream not found")
	}
	if err != nil {
		return nil, err
	}
	return &pb.MessageGetStreamStatisticsOutput{ID: input.ID, Statistics: stats}, nil
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
	for {
		count, offset, err := s.store.GetRange(input.StreamID, input.ShardID, input.Offset, buf)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			waitCh := make(chan *pb.MessagesStateTransitionMessagePut)
			handler := func(notification *pb.MessagesStateTransitionMessagePut) {
				waitCh <- notification
			}
			err := s.eventBus.SubscribeOnce(fmt.Sprintf("%s/message_put", input.ShardID), handler)
			if err != nil {
				close(waitCh)
				return nil, err
			}
			defer func() {
				s.eventBus.Unsubscribe(fmt.Sprintf("%s/message_put", input.ShardID), handler)
				close(waitCh)
			}()
			select {
			case <-waitCh:
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(30 * time.Second):
				return &pb.MessageGetMessagesOutput{
					Messages:   buf[:count],
					NextOffset: input.Offset,
					StreamID:   input.StreamID,
					ShardID:    input.ShardID,
				}, nil
			}
		}
		return &pb.MessageGetMessagesOutput{
			Messages:   buf[:count],
			NextOffset: offset,
			StreamID:   input.StreamID,
			ShardID:    input.ShardID,
		}, nil
	}
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
