package raft

import (
	"context"
	"crypto/sha1"
	"errors"
	fmt "fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	discovery "github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"go.uber.org/zap"
)

type statusChecker interface {
	getNodeStatus(peer string) string
	setStatus(string)
	getMembers() ([]*discovery.NodeService, error)
}

func waitForToken(ctx context.Context, token string, expectNodeCount int, s statusChecker, ticker *time.Ticker, logger *zap.Logger) error {
	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()
	now := time.Now()
	s.setStatus(fmt.Sprintf("%s-%s", raftStatusBootstrapping, token))
	prefix := fmt.Sprintf("%s-", raftStatusBootstrapping)
	lastCount := 0
	for {
		members, err := s.getMembers()
		if err != nil {
			return err
		}
		bootstrappingMembers := make([]*discovery.NodeService, 0)
		for _, member := range members {
			status := discovery.GetTagValue("raft_bootstrap_status", member.Tags)
			if status == raftStatusBootstrapped {
				return ErrBootstrappedNodeFound
			}
			if strings.TrimPrefix(status, prefix) == token {
				bootstrappingMembers = append(bootstrappingMembers, member)
			}
			count := len(bootstrappingMembers)
			if count == expectNodeCount {
				sortMembers(bootstrappingMembers)
				if token == SyncToken(bootstrappingMembers) {
					logger.Debug("members synchronization done", zap.Int("member_count", count), zap.Duration("member_synchronization_duration", time.Since(now)))
					return nil
				}
			} else {
				if count != lastCount {
					logger.Debug("synchronizing members", zap.Int("current_member_count", count))
				}
			}
			lastCount = count
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

func SyncToken(members []*discovery.NodeService) string {
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

func sortMembers(members []*discovery.NodeService) {
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
		err := s.joinCluster(ctx, name, expectNodeCount, s.logger, func(members []*discovery.NodeService) error {
			configuration := raft.Configuration{
				Servers: []raft.Server{
					{
						ID:      raft.ServerID(s.id),
						Address: raft.ServerAddress(fmt.Sprintf("127.0.0.1:%d", s.raftService.BindPort())),
					},
				},
			}
			err := s.raft.BootstrapCluster(configuration).Error()
			if err != nil {
				return err
			}
			s.logger.Debug("raft cluster bootstrapped")
			s.setStatus(raftStatusBootstrapped)
			s.cancelJoin()
			return nil
		})
		ch <- err
	}()
	return ch
}
func (s *raftlayer) joinCluster(ctx context.Context, name string, expectNodeCount int, logger *zap.Logger, done func(members []*discovery.NodeService) error) error {
	ticker := time.NewTicker(3 * time.Second)
	var members []*discovery.NodeService
	var err error
	start := time.Now()
	s.setStatus(raftStatusBootstrapping)
	for {
		members, err = s.getMembers()
		if err != nil {
			return err
		}
		bootstrappingMembers := make([]*discovery.NodeService, 0)
		for _, member := range members {
			status := discovery.GetTagValue("raft_bootstrap_status", member.Tags)
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
	logger.Debug("discovered members", zap.Duration("member_discovery_duration", time.Since(start)))
	sortMembers(members)
	err = waitForToken(ctx, SyncToken(members), expectNodeCount, s, ticker, logger)
	if err != nil {
		return err
	}
	if members[0].Peer != s.id {
		return ErrBootstrapRaceLost
	}
	return done(members)
}
