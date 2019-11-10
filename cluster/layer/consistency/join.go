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

func waitForToken(ctx context.Context, token string, expectNodeCount int, s statusChecker, ticker *time.Ticker, logger *zap.Logger) error {
	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()
	s.setStatus(fmt.Sprintf("%s-%s", raftStatusBootstrapping, token))
	prefix := fmt.Sprintf("%s-", raftStatusBootstrapping)
	for {
		members, err := s.getMembers()
		if err != nil {
			return err
		}
		bootstrappingMembers := make([]*pb.NodeService, 0)
		for _, member := range members {
			status := pb.GetTagValue("raft_bootstrap_status", member.Tags)
			if status == raftStatusBootstrapped {
				return ErrBootstrappedNodeFound
			}
			if strings.TrimPrefix(status, prefix) == token {
				bootstrappingMembers = append(bootstrappingMembers, member)
			}

			if len(bootstrappingMembers) == expectNodeCount {
				sortMembers(bootstrappingMembers)
				if token == SyncToken(bootstrappingMembers) {
					logger.Debug("members synchronization done", zap.Int("member_count", len(bootstrappingMembers)))
					return nil
				}
			} else {
				logger.Debug("synchronizing members", zap.Int("current_member_count", len(bootstrappingMembers)))
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			case <-timeout.C:
				return errors.New("members synchronization timed out")
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
		err := s.joinCluster(ctx, name, expectNodeCount, s.logger, func(members []*pb.NodeService) error {
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
		ch <- err
	}()
	return ch
}
func (s *raftlayer) joinCluster(ctx context.Context, name string, expectNodeCount int, logger *zap.Logger, done func(members []*pb.NodeService) error) error {
	ticker := time.NewTicker(3 * time.Second)
	var members []*pb.NodeService
	var err error
	s.logger.Debug("discovering members", zap.Int("expected_count", expectNodeCount))
	s.setStatus(raftStatusBootstrapping)
	for {
		members, err = s.discovery.Peers().EndpointsByService(fmt.Sprintf("%s_cluster", name))
		if err != nil {
			return err
		}
		bootstrappingMembers := make([]*pb.NodeService, 0)
		for _, member := range members {
			status := s.getNodeStatus(member.Peer)
			if strings.HasPrefix(status, raftStatusBootstrapping) {
				bootstrappingMembers = append(bootstrappingMembers, member)
			}
			if status == raftStatusBootstrapped {
				return ErrBootstrappedNodeFound
			}
		}
		if len(bootstrappingMembers) == expectNodeCount {
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
	err = waitForToken(ctx, SyncToken(members), expectNodeCount, s, ticker, logger)
	if err != nil {
		return err
	}
	if members[0].Peer != s.id {
		return nil
	}
	return done(members)
}
