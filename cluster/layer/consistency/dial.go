package consistency

import (
	"github.com/vx-labs/mqtt-broker/cluster/pb"
)

func (s *raftlayer) DialLeader(f func(pb.LayerClient) error) error {
	return f(s.leaderRPC)
}
