package layer

import (
	"context"
	fmt "fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
)

type raftlayer struct {
	id              string
	name            string
	raft            *raft.Raft
	selfRaftAddress raft.ServerAddress
	state           types.RaftState
	config          config.Config
	logger          *zap.Logger
	discovery       DiscoveryProvider
}
type DiscoveryProvider interface {
	EndpointsByService(name string) ([]*pb.NodeService, error)
}

func NewRaftLayer(logger *zap.Logger, userConfig config.Config, discovery DiscoveryProvider) (*raftlayer, error) {
	self := &raftlayer{
		id:              userConfig.ID,
		selfRaftAddress: raft.ServerAddress(fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, userConfig.AdvertisePort)),
		discovery:       discovery,
		logger:          logger,
		config:          userConfig,
	}
	return self, nil
}
func (s *raftlayer) Start(name string, state types.RaftState) error {
	s.name = name
	raftConfig := raft.DefaultConfig()
	s.state = state
	if os.Getenv("ENABLE_RAFT_LOG") == "false" {
		raftConfig.LogOutput = ioutil.Discard
	}
	raftConfig.LocalID = raft.ServerID(s.config.ID)
	raftBind := fmt.Sprintf("0.0.0.0:%d", s.config.BindPort)
	raftAdv := fmt.Sprintf("%s:%d", s.config.AdvertiseAddr, s.config.AdvertisePort)
	addr, err := net.ResolveTCPAddr("tcp", raftAdv)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(raftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots := raft.NewInmemSnapshotStore()
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	ra, err := raft.NewRaft(raftConfig, s, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra
	index := s.raft.LastIndex()
	if index != 0 {
		return nil
	}
	// hardcoded for now
	return s.joinCluster(name, 3)
}

func (s *raftlayer) joinCluster(name string, expectNodeCount int) error {
	// TODO check for existing cluster
	ticker := time.NewTicker(5 * time.Second)
	var members []*pb.NodeService
	var err error
	for {
		members, err = s.discovery.EndpointsByService(fmt.Sprintf("%s_cluster", name))
		if err != nil {
			return err
		}
		if len(members) == expectNodeCount {
			s.logger.Info("found other members", zap.Int("member_count", len(members)))
			break
		}
		s.logger.Info("waiting for other members to appear")
		<-ticker.C
	}
	ticker.Stop()
	sort.SliceStable(members, func(i, j int) bool {
		return strings.Compare(members[i].Peer, members[j].Peer) == -1
	})
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
	go s.leaderRoutine()
	err = s.raft.BootstrapCluster(configuration).Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *raftlayer) Status(ctx context.Context, input *pb.StatusInput) (*pb.StatusOutput, error) {
	return &pb.StatusOutput{
		Layer:  "raft",
		Status: "ok",
	}, nil
}

func (s *raftlayer) leaderRoutine() {
	for leader := range s.raft.LeaderCh() {
		if !leader {
			continue
		}
		members, err := s.discovery.EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
		if err != nil {
			s.logger.Error("failed to discover nodes", zap.Error(err))
			continue
		}
		for _, member := range members {
			if member.Peer != s.id {
				s.raft.AddVoter(raft.ServerID(member.Peer), raft.ServerAddress(member.NetworkAddress), 0, 100*time.Millisecond)
			}
		}
	}
}
func (s *raftlayer) Apply(log *raft.Log) interface{} {
	leader := s.raft.Leader() == s.selfRaftAddress
	return s.state.Apply(log.Data, leader)
}

type snapshot struct {
	src io.Reader
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := io.Copy(sink, s.src)
	return err
}

func (s *snapshot) Release() {}

func (s *raftlayer) Snapshot() (raft.FSMSnapshot, error) {
	src := s.state.Snapshot()
	return &snapshot{
		src: src,
	}, nil
}
func (s *raftlayer) Restore(snap io.ReadCloser) error {
	defer snap.Close()
	return s.state.Restore(snap)
}

func (s *raftlayer) Shutdown() error {
	return s.raft.Shutdown().Error()
}
func (s *raftlayer) ApplyEvent(event []byte) error {
	promise := s.raft.Apply(event, 300*time.Millisecond)
	err := promise.Error()
	if err != nil {
		return err
	}
	resp := promise.Response()
	if resp != nil {
		return resp.(error)
	}
	return nil
}
