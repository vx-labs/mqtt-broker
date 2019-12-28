package consistency

import (
	"context"
	"errors"
	fmt "fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	raftStatusInit          = "init"
	raftStatusBootstrapping = "bootstraping"
	raftStatusBootstrapped  = "bootstrapped"
)

var (
	ErrBootstrappedNodeFound = errors.New("bootstrapped node found")
	ErrBootstrapRaceLost     = errors.New("bootstrapped race lost")
)

type raftlayer struct {
	id              string
	name            string
	raft            *raft.Raft
	raftNetwork     *raft.NetworkTransport
	selfRaftAddress raft.ServerAddress
	state           types.RaftState
	config          config.Config
	logger          *zap.Logger
	discovery       DiscoveryProvider
	cancelJoin      context.CancelFunc
	leaderRPC       pb.LayerClient
	observer        *raft.Observer
	observations    chan raft.Observation
}
type DiscoveryProvider interface {
	UnregisterService(id string) error
	RegisterService(id string, address string) error
	AddServiceTag(service, key, value string) error
	RemoveServiceTag(name string, tag string) error
	DialService(id string) (*grpc.ClientConn, error)
	EndpointsByService(name string) ([]*pb.NodeService, error)
}

func New(logger *zap.Logger, userConfig config.Config, discovery DiscoveryProvider) (*raftlayer, error) {
	self := &raftlayer{
		id:              userConfig.ID,
		selfRaftAddress: raft.ServerAddress(fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, userConfig.AdvertisePort)),
		discovery:       discovery,
		logger:          logger,
		config:          userConfig,
	}
	return self, nil
}
func (s *raftlayer) joinExistingCluster(ctx context.Context) {
	defer s.cancelJoin()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		_, err := s.leaderRPC.RequestAdoption(ctx, &pb.RequestAdoptionInput{
			ID:      s.id,
			Address: fmt.Sprintf("%s:%d", s.config.AdvertiseAddr, s.config.AdvertisePort),
		})
		if err == nil {
			s.setStatus(raftStatusBootstrapped)
			return
		}
		s.logger.Error("failed to request adoption, retrying", zap.Error(err))
		<-ticker.C
		continue
	}
}
func (s *raftlayer) Start(name string, state types.RaftState) error {
	s.name = name
	leaderConn, err := s.discovery.DialService(fmt.Sprintf("%s_rpc?raft_status=leader", name))
	if err != nil {
		panic(err)
	}
	s.leaderRPC = pb.NewLayerClient(leaderConn)
	raftConfig := raft.DefaultConfig()
	s.state = state
	s.setStatus(raftStatusInit)
	// if os.Getenv("ENABLE_RAFT_LOG") != "true" {
	// 	raftConfig.LogOutput = ioutil.Discard
	// }
	raftConfig.LocalID = raft.ServerID(s.config.ID)
	raftBind := fmt.Sprintf("0.0.0.0:%d", s.config.BindPort)
	raftAdv := fmt.Sprintf("%s:%d", s.config.AdvertiseAddr, s.config.AdvertisePort)
	addr, err := net.ResolveTCPAddr("tcp", raftAdv)
	if err != nil {
		return err
	}
	var transport *raft.NetworkTransport
	if os.Getenv("ENABLE_RAFT_LOG") != "true" {
		transport, err = raft.NewTCPTransport(raftBind, addr, 5, 15*time.Second, ioutil.Discard)
	} else {
		transport, err = raft.NewTCPTransport(raftBind, addr, 5, 15*time.Second, os.Stderr)
	}
	if err != nil {
		return err
	}
	s.raftNetwork = transport
	// Create the snapshot store. This allows the Raft to truncate the log.
	retries := 3
	var snapshots raft.SnapshotStore
	for {
		snapshots, err = raft.NewFileSnapshotStore(buildDataDir(s.config.ID), 5, os.Stderr)
		retries--
		if err == nil {
			break
		}
		if retries == 0 {
			return fmt.Errorf("failed to create snapshot store: %s", err)
		}
	}
	filename := fmt.Sprintf("%s/raft-%s.db", buildDataDir(s.config.ID), s.name)
	boltDB, err := raftboltdb.NewBoltStore(filename)
	if err != nil {
		return fmt.Errorf("failed to create boltdb store: %s", err)
	}
	cacheStore, err := raft.NewLogCache(512, boltDB)
	if err != nil {
		boltDB.Close()
		return err
	}
	logStore := cacheStore
	stableStore := boltDB
	ra, err := raft.NewRaft(raftConfig, s, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelJoin = cancel
	go s.leaderRoutine()
	index := s.raft.LastIndex()
	err = s.discovery.RegisterService(fmt.Sprintf("%s_cluster", name), fmt.Sprintf("%s:%d", s.config.AdvertiseAddr, s.config.AdvertisePort))
	if err != nil {
		return err
	}
	if index != 0 {
		s.joinExistingCluster(ctx)
		return nil
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.logger.Debug("join session canceled")
				return
			case err := <-s.startClusterJoin(ctx, name, 3):
				if err != nil {
					if err == ErrBootstrappedNodeFound || err == ErrBootstrapRaceLost {
						s.joinExistingCluster(ctx)
						return
					}
					s.logger.Warn("failed to join cluster, retrying", zap.Error(err))
				}
			}
		}
	}()
	return nil
}

func (s *raftlayer) logRaftStatus(log *raft.Log) {
	//s.logger.Debug("raft status", zap.Uint64("raft_last_index", s.raft.LastIndex()), zap.Uint64("raft_applied_index", s.raft.AppliedIndex()), zap.Uint64("raft_current_index", log.Index))
}
func (s *raftlayer) Health() string {
	if s.raft == nil {
		return "critical"
	}
	if s.raft.AppliedIndex() == 0 {
		return "warning"
	}
	if (s.raft.Leader()) == "" {
		return "critical"
	}
	return "ok"
}

func (s *raftlayer) discoveredMembers() []string {
	members := []string{}
	discovered, err := s.discovery.EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
	if err != nil {
		return members
	}

	for _, member := range discovered {
		members = append(members, member.Peer)
	}
	return members
}
func (s *raftlayer) raftMembers() []string {
	members := []string{}
	cProm := s.raft.GetConfiguration()
	err := cProm.Error()
	if err != nil {
		return members
	}
	clusterConfig := cProm.Configuration()
	for _, server := range clusterConfig.Servers {
		members = append(members, string(server.ID))
	}
	return members
}

func (s *raftlayer) IsLeader() bool {
	return s.raft != nil && s.raft.State() == raft.Leader
}
func (s *raftlayer) isNodeAdopted(id string) bool {
	cProm := s.raft.GetConfiguration()
	err := cProm.Error()
	if err != nil {
		return false
	}
	clusterConfig := cProm.Configuration()
	for _, server := range clusterConfig.Servers {
		if string(server.ID) == id {
			return true
		}
	}
	return false
}
func (s *raftlayer) removeMember(id string) {
	err := s.raft.RemoveServer(raft.ServerID(id), 0, 0).Error()
	if err != nil {
		s.logger.Error("failed to remove dead node", zap.Error(err))
		return
	}
	s.logger.Info("removed dead node")
}
func (s *raftlayer) addMember(id, address string) {
	err := s.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(address), 0, 0).Error()
	if err != nil {
		s.logger.Error("failed to add new node", zap.String("new_node", id[:8]), zap.Error(err))
		return
	}
	s.logger.Info("adopted new raft node", zap.String("new_node", id[:8]))
}
func (s *raftlayer) syncMembers() {
	members, err := s.discovery.EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
	if err != nil {
		s.logger.Error("failed to discover nodes", zap.Error(err))
		panic(err)
	}
	cProm := s.raft.GetConfiguration()
	err = cProm.Error()
	if err != nil {
		panic(err)
	}
	clusterConfig := cProm.Configuration()
	for idx := range cProm.Configuration().Servers {
		server := clusterConfig.Servers[idx]
		found := false
		for _, member := range members {
			if string(server.ID) == member.Peer {
				found = true
			}
		}
		if !found {
			s.removeMember(string(server.ID))
		}
	}
}
func (s *raftlayer) leaderRoutine() {
	s.observations = make(chan raft.Observation)
	s.observer = raft.NewObserver(s.observations, true, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.LeaderObservation)
		return ok
	})
	s.raft.RegisterObserver(s.observer)
	for range s.observations {
		leader := s.IsLeader()
		if !leader {
			s.discovery.RemoveServiceTag(s.name, "raft_status")
			s.discovery.RemoveServiceTag(fmt.Sprintf("%s_rpc", s.name), "raft_status")
			continue
		}
		s.logger.Info("raft cluster leadership acquired")
		s.discovery.AddServiceTag(s.name, "raft_status", "leader")
		s.discovery.AddServiceTag(fmt.Sprintf("%s_rpc", s.name), "raft_status", "leader")
	}
}
func (s *raftlayer) Apply(log *raft.Log) interface{} {
	if log.Type == raft.LogCommand {
		return s.state.Apply(log.Data)
	}
	return nil
}

type snapshot struct {
	src io.ReadCloser
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := io.Copy(sink, s.src)
	if err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *snapshot) Release() {
	s.src.Close()
}

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

func (s *raftlayer) ApplyEvent(event []byte) error {
	retries := 5
	for retries > 0 {
		retries--
		if !s.IsLeader() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.logger.Debug("forwarding event to leader")
			return s.DialLeader(func(client pb.LayerClient) error {
				_, err := client.SendEvent(ctx, &pb.SendEventInput{Payload: event})
				return err
			})
		}
		promise := s.raft.Apply(event, 30*time.Second)
		err := promise.Error()
		if err != nil {
			if err == raft.ErrLeadershipLost || err == raft.ErrLeadershipTransferInProgress {
				s.logger.Info("leadership lost while committing log, retrying")
				continue
			}
			return err
		}
		resp := promise.Response()
		if resp != nil {
			return resp.(error)
		}
		return nil
	}
	return errors.New("failed to commit to leader after 5 retries")
}

func (s *raftlayer) setStatus(status string) {
	err := s.discovery.AddServiceTag(fmt.Sprintf("%s_cluster", s.name), "raft_bootstrap_status", status)
	if err != nil {
		s.logger.Warn("failed to set raft bootstrap tag", zap.Error(err))
	}
}
func (s *raftlayer) getMembers() ([]*pb.NodeService, error) {
	return s.discovery.EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
}
func (s *raftlayer) getNodeStatus(peer string) string {
	discovered, err := s.discovery.EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
	if err != nil {
		return ""
	}
	for _, service := range discovered {
		if service.Peer == peer {
			return pb.GetTagValue("raft_bootstrap_status", service.Tags)
		}
	}
	return ""
}
