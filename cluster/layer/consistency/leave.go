package consistency

import (
	fmt "fmt"
	"time"

	"go.uber.org/zap"
)

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
		err := s.raftNetwork.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *raftlayer) Leave() error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	err := s.discovery.UnregisterService(fmt.Sprintf("%s_cluster", s.name))
	if err != nil {
		s.logger.Error("failed to unregister raft service from discovery", zap.Error(err))
	}
	s.logger.Info("unregistered service from discovery")
	numPeers := len(s.raftMembers())
	isLeader := s.IsLeader()
	if isLeader && numPeers > 1 {
		if b := s.raft.Barrier(5 * time.Second); b.Error() != nil {
			s.logger.Error("failed to wait for other members to catch-up log", zap.Error(err))
			return err
		}
		err = s.raft.LeadershipTransfer().Error()
		if err != nil {
			s.logger.Error("failed to transfert raft cluster", zap.Error(err))
			return err
		}
		for {
			isLeader = s.IsLeader()
			if !isLeader {
				break
			}
			<-ticker.C
		}
		s.logger.Info("raft leadership transfered")
	}

	if !isLeader {
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
					break
				}
			}
		}
		if !left {
			s.logger.Warn("failed to leave raft configuration gracefully, timeout")
		} else {
			s.logger.Info("raft cluster left")
		}
	}
	return nil
}
