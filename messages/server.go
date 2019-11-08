package queues

import (
	"context"
	fmt "fmt"
	"net"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/messages/pb"
	"github.com/vx-labs/mqtt-broker/messages/store"
	"go.uber.org/zap"
)

const (
	dbPath = "/var/tmp/mqtt-broker-messages"
)

type server struct {
	id        string
	store     *store.BoltStore
	state     types.RaftServiceLayer
	ctx       context.Context
	listeners []net.Listener
	logger    *zap.Logger
}

func New(id string, logger *zap.Logger) *server {
	logger.Debug("opening messages durable store")
	boltstore, err := store.New(store.Options{
		NoSync: false,
		Path:   fmt.Sprintf("%s-%s-messages.bolt", dbPath, id),
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

func (s *server) PutMessage(ctx context.Context, input *pb.MessagePutMessageInput) (*pb.MessagePutMessageOutput, error) {
	// FIXME: do not use time as a key generator
	idx := time.Now().UnixNano()
	payload, err := proto.Marshal(input.Publish)
	if err != nil {
		s.logger.Error("failed to encode message", zap.Error(err))
		return nil, err
	}
	err = s.commitEvent(&pb.MessagesStateTransition{
		Kind: MessagePut,
		MessagePut: &pb.MessagesStateTransitionMessagePut{
			Offset:  uint64(idx),
			Payload: payload,
			Tenant:  input.Tenant,
		},
	})
	if err != nil {
		s.logger.Error("failed to commit queue put event", zap.Error(err))
	}
	return &pb.MessagePutMessageOutput{}, err
}
func (s *server) PutMessageBatch(ctx context.Context, ev *pb.MessagePutMessageBatchInput) (*pb.MessagePutMessageBatchOutput, error) {
	// FIXME: do not use time as a key generator
	now := time.Now().UnixNano()
	event := []*pb.MessagesStateTransition{}
	idx := now
	for _, input := range ev.Batches {
		payload, err := proto.Marshal(input.Publish)
		if err != nil {
			s.logger.Error("failed to encode message", zap.Error(err))
			return nil, err
		}
		event = append(event, &pb.MessagesStateTransition{
			Kind: MessagePut,
			MessagePut: &pb.MessagesStateTransitionMessagePut{
				Offset:  uint64(idx),
				Payload: payload,
				Tenant:  input.Tenant,
			},
		})
		idx++
	}
	err := s.commitEvent(event...)
	if err != nil {
		s.logger.Error("failed to commit queue put batch event", zap.Error(err))
	}
	return &pb.MessagePutMessageBatchOutput{}, err
}
