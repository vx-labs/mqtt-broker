package raft

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	discovery "github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"go.uber.org/zap"
)

type statusChecker interface {
	getMembers() ([]*discovery.NodeService, error)
}

func sortMembers(members []*discovery.NodeService) {
	sort.SliceStable(members, func(i, j int) bool {
		return strings.Compare(members[i].ID, members[j].ID) == -1
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
				Servers: []raft.Server{},
			}
			for idx := range members {
				member := members[idx]
				configuration.Servers = append(configuration.Servers, raft.Server{
					ID:       raft.ServerID(member.ID),
					Suffrage: raft.Voter,
					Address:  raft.ServerAddress(member.NetworkAddress),
				})
			}
			err := s.raft.BootstrapCluster(configuration).Error()
			if err != nil {
				return err
			}
			s.logger.Debug("raft cluster bootstrapped")
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
	for {
		members, err = s.getMembers()
		if err != nil {
			return err
		}
		bootstrappingMembers := make([]*discovery.NodeService, 0)
		for _, member := range members {
			if member.Health == "passing" {
				return ErrBootstrappedNodeFound
			}
			if member.Health == "warning" {
				bootstrappingMembers = append(bootstrappingMembers, member)
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
	return done(members)
}
