package raft

import (
	"context"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/adapters/cp/pb"
	"go.uber.org/zap"
)

func (s *raftlayer) Shutdown() error {
	if s.raft == nil {
		return nil
	}
	s.logger.Debug("shutting down raft layer")
	err := s.Leave()
	if err != nil {
		s.logger.Error("failed to leave raft cluster", zap.Error(err))
		return err
	}
	s.grpcServer.GracefulStop()
	if s.raftNetwork != nil {
		err := s.raftNetwork.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *raftlayer) Leave() error {
	var err error
	s.raft.DeregisterObserver(s.observer)
	close(s.observations)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	numPeers := len(s.raftMembers())
	isLeader := s.raft.VerifyLeader().Error() == nil
leaderLogic:
	if isLeader && numPeers > 1 {
		s.logger.Info("waiting for other members to catch-up raft log")
		if b := s.raft.Barrier(0); b.Error() != nil {
			s.logger.Error("failed to wait for other members to catch-up raft log", zap.Error(b.Error()))
			return b.Error()
		}
		err := s.raft.RemoveServer(raft.ServerID(s.id), 0, 0).Error()
		if err == nil {
			s.logger.Info("removed ourselves from cluster")
		}
		s.logger.Info("raft leadership transfered")
	}

	if !isLeader {
		for {
			_, err = s.leaderRPC.PrepareShutdown(ctx, &pb.PrepareShutdownInput{
				ID:    s.id,
				Index: 0,
			})
			if err == nil {
				break
			}
			s.logger.Warn("failed to ask leader to remove us from cluster, retrying", zap.Error(err))
			isLeader = s.raft.VerifyLeader().Error() == nil
			if isLeader {
				goto leaderLogic
			}
			<-ticker.C
		}
		left := false
		deadline := time.Now().Add(15 * time.Second)
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
				}
			}
			if left {
				break
			}
		}
		if !left {
			s.logger.Warn("failed to leave raft cluster")
		} else {
			s.logger.Debug("raft cluster left")
		}
	}
	err = s.rpcListener.Close()
	if err != nil {
		s.logger.Error("failed to close raft rpc service", zap.Error(err))
	}
	err = s.raftNetwork.Close()
	if err != nil {
		s.logger.Error("failed to close raft service", zap.Error(err))
	}
	s.logger.Info("unregistered service from discovery")
	return s.raft.Shutdown().Error()
}
