package consistency

import (
	fmt "fmt"
	"time"

	"github.com/hashicorp/raft"
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
	err := s.discovery.UnregisterService(fmt.Sprintf("%s_cluster", s.name))
	if err != nil {
		s.logger.Error("failed to unregister raft service from discovery", zap.Error(err))
	}
	s.logger.Info("unregistered service from discovery")
	numPeers := len(s.raftMembers())
	isLeader := s.IsLeader()
	if isLeader && numPeers > 1 {
		err := s.raft.RemoveServer(raft.ServerID(s.id), 0, 0).Error()
		if err != nil {
			s.logger.Error("failed to leave raft cluster", zap.Error(err))
			return err
		}
		s.logger.Info("raft leadership transfered")
	}

	if !isLeader {
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
		} else {
			s.logger.Info("raft cluster left")
		}
	}
	return nil
}
