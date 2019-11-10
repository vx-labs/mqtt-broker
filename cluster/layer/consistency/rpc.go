package consistency

import (
	"context"
	"errors"

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
	err := s.raft.RemoveServer(raft.ServerID(input.ID), input.Index, 0).Error()
	if err == nil {
		s.logger.Info("removed leaving node", zap.Strings("raft_members", s.raftMembers()), zap.Strings("discovered_members", s.discoveredMembers()))
	}
	return &pb.PrepareShutdownOutput{}, err
}
