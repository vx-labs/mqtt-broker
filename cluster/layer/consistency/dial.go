package consistency

import (
	"errors"
	fmt "fmt"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (s *raftlayer) DialLeader(f func(pb.LayerClient) error) error {
	leader := string(s.raft.Leader())
	serviceRPC := fmt.Sprintf("%s_rpc", s.name)
	members, err := s.discovery.Peers().EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
	if err != nil {
		s.logger.Error("failed to discover nodes", zap.Error(err))
		return err
	}
	for _, member := range members {
		if member.NetworkAddress == leader {
			return s.discovery.DialAddress(serviceRPC, member.Peer, func(c *grpc.ClientConn) error {
				client := pb.NewLayerClient(c)
				return f(client)
			})
		}
	}
	s.logger.Error("failed to find leader addess", zap.String("leader_address", leader))
	return errors.New("failed to find leader address")
}
