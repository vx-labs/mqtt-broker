package consistency

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"go.uber.org/zap"
)

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

func (s *raftlayer) Shutdown() error {
	if s.raft == nil {
		return nil
	}
	s.logger.Info("shuting down raft layer")
	err := s.Leave()
	if err != nil {
		return err
	}
	if s.raftNetwork != nil {
		return s.raftNetwork.Close()
	}
	return s.raft.Shutdown().Error()
}
func (s *raftlayer) Leave() error {
	ctx := context.Background()
	numPeers := len(s.raftMembers())

	if s.IsLeader() {
		if numPeers > 1 {
			s.logger.Info("transfering raft leadership")
			err := s.raft.RemoveServer(raft.ServerID(s.id), 0, 0).Error()
			if err != nil {
				s.logger.Error("failed to leave raft cluster", zap.Error(err))
			}
		}
	} else {
		s.logger.Info("leaving raft cluster")
		err := s.DialLeader(func(client pb.LayerClient) error {
			_, err := client.PrepareShutdown(ctx, &pb.PrepareShutdownInput{
				ID:    s.id,
				Index: 0,
			})
			return err
		})
		if err != nil {
			s.logger.Error("failed to ask leader to remove use from cluster", zap.Error(err))
		}
		left := false
		deadline := time.Now().Add(15 * time.Second)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for !left && time.Now().Before(deadline) {
			<-ticker.C
			future := s.raft.GetConfiguration()
			if err := future.Error(); err != nil {
				s.logger.Error("failed to get raft configuration", zap.Error(err))
				break
			}
			left = true
			for _, server := range future.Configuration().Servers {
				if server.Address == s.selfRaftAddress {
					left = false
					break
				}
			}
		}
		if !left {
			s.logger.Warn("failed to leave raft configuration gracefully, timeout")
		}
	}
	return nil
}
