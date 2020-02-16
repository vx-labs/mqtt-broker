package raft

import (
	"context"
	"errors"
	fmt "fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/adapters/cp/pb"
	discovery "github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"go.uber.org/zap"
)

var (
	ErrBootstrappedNodeFound = errors.New("bootstrapped node found")
	ErrBootstrapRaceLost     = errors.New("bootstrapped race lost")
)

type raftlayer struct {
	id              string
	name            string
	userService     Service
	raft            *raft.Raft
	raftNetwork     *raft.NetworkTransport
	selfRaftAddress raft.ServerAddress
	state           pb.CPState
	logger          *zap.Logger
	raftService     Service
	rpcService      Service
	cancelJoin      context.CancelFunc
	leaderRPC       pb.LayerClient
	observer        *raft.Observer
	observations    chan raft.Observation
	wasLeader       bool
	grpcServer      *grpc.Server
	rpcListener     net.Listener
}
type Service interface {
	DiscoverEndpoints() ([]*discovery.NodeService, error)
	Address() string
	ID() string
	Name() string
	Dial() (*grpc.ClientConn, error)
	ListenTCP() (net.Listener, error)
}

func NewRaftSynchronizer(id string, userService Service, service Service, rpc Service, state pb.CPState, logger *zap.Logger) *raftlayer {
	self := &raftlayer{
		id:              service.ID(),
		userService:     userService,
		name:            service.Name(),
		selfRaftAddress: raft.ServerAddress(service.Address()),
		raftService:     service,
		rpcService:      rpc,
		logger:          logger,
		state:           state,
	}
	err := self.Serve()
	if err != nil {
		panic(err)
	}
	err = self.start()
	if err != nil {
		panic(err)
	}
	return self
}
func (s *raftlayer) joinExistingCluster(ctx context.Context) {
	s.logger.Info("joining raft cluster")
	defer s.cancelJoin()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	retries := 100
	var err error
	for retries > 0 {
		retries--
		_, err = s.leaderRPC.RequestAdoption(ctx, &pb.RequestAdoptionInput{
			ID:      s.id,
			Address: s.raftService.Address(),
		})
		if err == nil {
			s.logger.Info("joined raft cluster")
			return
		}
		<-ticker.C
		continue
	}
	s.logger.Error("failed to request adoption", zap.Error(err))
}
func (s *raftlayer) start() error {
	leaderConn, err := s.rpcService.Dial()
	if err != nil {
		panic(err)
	}
	s.leaderRPC = pb.NewLayerClient(leaderConn)
	raftConfig := raft.DefaultConfig()
	/*	if os.Getenv("ENABLE_RAFT_LOG") != "true" {
		raftConfig.LogOutput = ioutil.Discard
	}*/
	raftConfig.LocalID = raft.ServerID(s.id)
	listener, err := s.raftService.ListenTCP()
	if err != nil {
		return err
	}
	addr, err := net.ResolveTCPAddr("tcp", s.raftService.Address())
	if err != nil {
		return err
	}
	var transport *raft.NetworkTransport
	if os.Getenv("ENABLE_RAFT_LOG") != "true" {

		transport = raft.NewNetworkTransport(&TCPStreamLayer{
			listener:  listener,
			advertise: addr,
		}, 5, 15*time.Second, ioutil.Discard)
	} else {
		transport = raft.NewNetworkTransport(&TCPStreamLayer{
			listener:  listener,
			advertise: addr,
		}, 5, 15*time.Second, os.Stderr)
	}
	if err != nil {
		return err
	}
	s.raftNetwork = transport
	// Create the snapshot store. This allows the Raft to truncate the log.
	retries := 3
	var snapshots raft.SnapshotStore
	for {
		snapshots, err = raft.NewFileSnapshotStore(buildDataDir(s.id), 5, os.Stderr)
		retries--
		if err == nil {
			break
		}
		if retries == 0 {
			return fmt.Errorf("failed to create snapshot store: %s", err)
		}
	}
	filename := fmt.Sprintf("%s/raft-%s.db", buildDataDir(s.id), s.name)
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
	if index != 0 {
		if !s.IsLeader() {
			s.joinExistingCluster(ctx)
		}
		return nil
	}

	go func() {
		select {
		case <-ctx.Done():
			s.logger.Debug("join session canceled")
			return
		case err := <-s.startClusterJoin(ctx, s.name, 3):
			if err != nil {
				if err == ErrBootstrappedNodeFound || err == ErrBootstrapRaceLost {
					s.joinExistingCluster(ctx)
					return
				}
				s.logger.Warn("failed to join raft cluster, retrying", zap.Error(err))
			} else {
				return
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
	if s.IsLeader() {
		return "ok"
	}
	return "warning"
}

func (s *raftlayer) discoveredMembers() []string {
	members := []string{}
	discovered, err := s.raftService.DiscoverEndpoints()
	if err != nil {
		return members
	}

	for _, member := range discovered {
		members = append(members, member.ID)
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
	return s.raft != nil && s.raft.Leader() == s.selfRaftAddress
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
	members, err := s.raftService.DiscoverEndpoints()
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
			if string(server.ID) == member.ID {
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
	for ob := range s.observations {
		leader := ob.Raft.Leader()
		if leader == s.selfRaftAddress {
			s.logger.Info("raft cluster leadership acquired")
		} else if leader != "" && s.wasLeader {
			s.logger.Info("raft cluster leadership lost")
		}
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

func (s *raftlayer) getMembers() ([]*discovery.NodeService, error) {
	return s.raftService.DiscoverEndpoints()
}
