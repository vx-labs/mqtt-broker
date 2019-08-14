package layer

import (
	fmt "fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
)

type raftlayer struct {
	id              string
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
	raftConfig := raft.DefaultConfig()
	s.state = state
	if os.Getenv("ENABLE_RAFT_LOG") != "true" {
		raftConfig.LogOutput = ioutil.Discard
	}
	raftConfig.LocalID = raft.ServerID(s.config.ID)
	raftBind := fmt.Sprintf("0.0.0.0:%d", s.config.BindPort)
	addr, err := net.ResolveTCPAddr("tcp", raftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(raftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots := raft.NewInmemSnapshotStore()
	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore
	logStore = raft.NewInmemStore()
	stableStore = raft.NewInmemStore()

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(raftConfig, s, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra
	// hardcoded for now
	return s.joinCluster(name, 3)
}

func (s *raftlayer) joinCluster(name string, expectNodeCount int) error {
	// TODO check for existing cluster
	ticker := time.NewTicker(5 * time.Second)
	var members []*pb.NodeService
	var err error
	for {
		members, err = s.discovery.EndpointsByService(name)
		if err != nil {
			return err
		}
		if len(members) == expectNodeCount {
			break
		}
		<-ticker.C
	}
	ticker.Stop()
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(s.id),
				Address: raft.ServerAddress(fmt.Sprintf("127.0.0.1:%d", s.config.BindPort)),
			},
		},
	}
	for _, member := range members {
		configuration.Servers = append(configuration.Servers, raft.Server{
			ID:      raft.ServerID(member.ID),
			Address: raft.ServerAddress(member.NetworkAddress),
		})
	}
	return s.raft.BootstrapCluster(configuration).Error()
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
