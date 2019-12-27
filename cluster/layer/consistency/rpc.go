package consistency

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"go.uber.org/zap"
)

func (s *raftlayer) SendEvent(ctx context.Context, input *pb.SendEventInput) (*pb.SendEventOutput, error) {
	return &pb.SendEventOutput{}, s.ApplyEvent(input.Payload)
}

func (s *raftlayer) PrepareShutdown(ctx context.Context, input *pb.PrepareShutdownInput) (*pb.PrepareShutdownOutput, error) {
	if !s.IsLeader() {
		return nil, errors.New("current node is not leader")
	}
	if b := s.raft.Barrier(5 * time.Second); b.Error() != nil {
		s.logger.Error("failed to wait for other members to catch-up log", zap.Error(b.Error()))
		return nil, b.Error()
	}
	err := s.raft.RemoveServer(raft.ServerID(input.ID), input.Index, 0).Error()
	if err == nil {
		s.logger.Info("removed leaving node", zap.String("leaving_node", input.ID))
	}
	return &pb.PrepareShutdownOutput{}, err
}
func (s *raftlayer) RequestAdoption(ctx context.Context, input *pb.RequestAdoptionInput) (*pb.RequestAdoptionOutput, error) {
	if !s.IsLeader() {
		return nil, errors.New("current node is not leader")
	}
	s.syncMembers()
	err := s.raft.AddVoter(raft.ServerID(input.ID), raft.ServerAddress(input.Address), 0, 30*time.Second).Error()
	if err == nil {
		s.logger.Info("added new node", zap.String("new_node", input.ID))
	}
	return &pb.RequestAdoptionOutput{}, err
}
