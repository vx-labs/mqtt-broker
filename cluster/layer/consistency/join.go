package consistency

import (
	"crypto/sha1"
	"errors"
	fmt "fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"go.uber.org/zap"
)

func (s *raftlayer) waitForToken(token string, expectNodeCount int) error {
	ticker := time.NewTicker(1 * time.Second)
	timeout := time.NewTimer(60 * time.Second)
	defer ticker.Stop()
	defer timeout.Stop()
	s.status = fmt.Sprintf("%s-%s", raftStatusBootstrapping, token)
	prefix := fmt.Sprintf("%s-", raftStatusBootstrapping)
	for {
		members, err := s.discovery.Peers().EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
		if err != nil {
			return err
		}
		bootstrappingMembers := make([]*pb.NodeService, 0)
		s.logger.Info("waiting for other members to complete step 2", zap.String("sync_token", token))
		for _, member := range members {
			status := strings.TrimPrefix(s.nodeStatus(member.Peer, s.name), prefix)
			if status == raftStatusBootstrapped {
				return ErrBootstrappedNodeFound
			}
			if status == token {
				bootstrappingMembers = append(bootstrappingMembers, member)
			}

			if len(bootstrappingMembers) == expectNodeCount {
				sortMembers(bootstrappingMembers)
				if token == SyncToken(bootstrappingMembers) {
					s.logger.Info("members synchronization done", zap.String("sync_token", token), zap.Int("member_count", len(bootstrappingMembers)))
					return nil
				}
			}
			select {
			case <-ticker.C:
			case <-timeout.C:
				return errors.New("step 2 timed out")
			}
		}
	}
}

func SyncToken(members []*pb.NodeService) string {
	if !sort.SliceIsSorted(members, func(i, j int) bool {
		return strings.Compare(members[i].Peer, members[j].Peer) == -1
	}) {
		panic("members are not sorted")
	}
	bootstrapToken := sha1.New()
	for _, member := range members {
		bootstrapToken.Write([]byte(member.Peer))
	}
	return fmt.Sprintf("%x", (bootstrapToken.Sum(nil)))
}

func sortMembers(members []*pb.NodeService) {
	sort.SliceStable(members, func(i, j int) bool {
		return strings.Compare(members[i].Peer, members[j].Peer) == -1
	})
}

func (s *raftlayer) joinCluster(name string, expectNodeCount int) error {
	s.discovery.RegisterService(fmt.Sprintf("%s_cluster", name), fmt.Sprintf("%s:%d", s.config.AdvertiseAddr, s.config.AdvertisePort))
	s.status = raftStatusBootstrapping
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var members []*pb.NodeService
	var err error
	s.logger.Info("waiting for other cluster members to be discovered", zap.Int("expected_count", expectNodeCount))
	for {
		members, err = s.discovery.Peers().EndpointsByService(fmt.Sprintf("%s_cluster", name))
		if err != nil {
			return err
		}
		bootstrappingMembers := make([]*pb.NodeService, 0)
		for _, member := range members {
			status := s.nodeStatus(member.Peer, name)
			if strings.HasPrefix(status, raftStatusBootstrapping) {
				bootstrappingMembers = append(bootstrappingMembers, member)
			}
			if status == raftStatusBootstrapped {
				return ErrBootstrappedNodeFound
			}
		}
		if len(bootstrappingMembers) == expectNodeCount {
			s.logger.Info("found other members", zap.Int("member_count", len(bootstrappingMembers)), zap.Int("expected_count", expectNodeCount))
			members = bootstrappingMembers
			break
		}
		<-ticker.C
	}
	sortMembers(members)
	err = s.waitForToken(SyncToken(members), expectNodeCount)
	if err != nil {
		return err
	}
	if members[0].Peer != s.id {
		return nil
	}
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(s.id),
				Address: raft.ServerAddress(fmt.Sprintf("127.0.0.1:%d", s.config.BindPort)),
			},
		},
	}
	err = s.raft.BootstrapCluster(configuration).Error()
	if err != nil {
		return err
	}
	s.logger.Info("raft cluster bootstrapped")
	s.status = raftStatusBootstrapped
	return nil
}