package consistency

import (
	"context"
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

type statusChecker interface {
	getNodeStatus(peer string) string
	setStatus(string)
	getMembers() ([]*pb.NodeService, error)
}

func waitForToken(ctx context.Context, token string, expectNodeCount int, s statusChecker, logger *zap.Logger) error {
	ticker := time.NewTicker(1 * time.Second)
	timeout := time.NewTimer(60 * time.Second)
	defer ticker.Stop()
	defer timeout.Stop()
	s.setStatus(fmt.Sprintf("%s-%s", raftStatusBootstrapping, token))
	prefix := fmt.Sprintf("%s-", raftStatusBootstrapping)
	for {
		members, err := s.getMembers()
		if err != nil {
			return err
		}
		bootstrappingMembers := make([]*pb.NodeService, 0)
		logger.Debug("waiting for other members to complete step 2", zap.String("sync_token", token))
		for _, member := range members {
			status := strings.TrimPrefix(s.getNodeStatus(member.Peer), prefix)
			if status == raftStatusBootstrapped {
				return ErrBootstrappedNodeFound
			}
			if status == token {
				bootstrappingMembers = append(bootstrappingMembers, member)
			}

			if len(bootstrappingMembers) == expectNodeCount {
				sortMembers(bootstrappingMembers)
				if token == SyncToken(bootstrappingMembers) {
					logger.Debug("members synchronization done", zap.String("sync_token", token), zap.Int("member_count", len(bootstrappingMembers)))
					return nil
				}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
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

func (s *raftlayer) startClusterJoin(ctx context.Context, name string, expectNodeCount int) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		ch <- s.joinCluster(ctx, name, expectNodeCount, s.logger, func(members []*pb.NodeService) error {
			configuration := raft.Configuration{
				Servers: []raft.Server{
					{
						ID:      raft.ServerID(s.id),
						Address: raft.ServerAddress(fmt.Sprintf("127.0.0.1:%d", s.config.BindPort)),
					},
				},
			}
			err := s.raft.BootstrapCluster(configuration).Error()
			if err != nil {
				return err
			}
			s.logger.Info("raft cluster bootstrapped")
			s.status = raftStatusBootstrapped
			return nil
		})
	}()
	return ch
}
func (s *raftlayer) joinCluster(ctx context.Context, name string, expectNodeCount int, logger *zap.Logger, done func(members []*pb.NodeService) error) error {
	s.discovery.RegisterService(fmt.Sprintf("%s_cluster", name), fmt.Sprintf("%s:%d", s.config.AdvertiseAddr, s.config.AdvertisePort))
	s.status = raftStatusBootstrapping
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var members []*pb.NodeService
	var err error
	s.logger.Debug("waiting for other cluster members to be discovered", zap.Int("expected_count", expectNodeCount))
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
			s.logger.Debug("found other members", zap.Int("member_count", len(bootstrappingMembers)), zap.Int("expected_count", expectNodeCount))
			members = bootstrappingMembers
			break
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	sortMembers(members)
	err = waitForToken(ctx, SyncToken(members), expectNodeCount, s, logger)
	if err != nil {
		return err
	}
	if members[0].Peer != s.id {
		return nil
	}
	return done(members)
}
