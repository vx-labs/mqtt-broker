package raft

import (
	"github.com/vx-labs/mqtt-broker/adapters/cp/pb"
)

func (s *raftlayer) DialLeader(f func(pb.LayerClient) error) error {
	return f(s.leaderRPC)
}
