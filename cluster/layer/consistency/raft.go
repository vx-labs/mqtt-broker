package consistency

import (
	"context"
	"crypto/sha1"
	"errors"
	fmt "fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/vx-labs/mqtt-broker/cluster/peers"

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
)

type raftlayer struct {
	id              string
	name            string
	status          string
	raft            *raft.Raft
	selfRaftAddress raft.ServerAddress
	state           types.RaftState
	config          config.Config
	logger          *zap.Logger
	discovery       DiscoveryProvider
}
type DiscoveryProvider interface {
	DialAddress(service, id string, f func(*grpc.ClientConn) error) error
	Peers() peers.PeerStore
}

func New(logger *zap.Logger, userConfig config.Config, discovery DiscoveryProvider) (*raftlayer, error) {
	self := &raftlayer{
		id:              userConfig.ID,
		selfRaftAddress: raft.ServerAddress(fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, userConfig.AdvertisePort)),
		discovery:       discovery,
		logger:          logger.WithOptions(zap.Fields(zap.String("emitter", "raft_layer"))),
		config:          userConfig,
		status:          raftStatusInit,
	}
	return self, nil
}
func (s *raftlayer) Start(name string, state types.RaftState) error {
	s.name = name
	raftConfig := raft.DefaultConfig()
	s.state = state
	if os.Getenv("ENABLE_RAFT_LOG") != "true" {
		raftConfig.LogOutput = ioutil.Discard
	}
	raftConfig.LocalID = raft.ServerID(s.config.ID)
	raftBind := fmt.Sprintf("0.0.0.0:%d", s.config.BindPort)
	raftAdv := fmt.Sprintf("%s:%d", s.config.AdvertiseAddr, s.config.AdvertisePort)
	addr, err := net.ResolveTCPAddr("tcp", raftAdv)
	if err != nil {
		return err
	}
	var transport *raft.NetworkTransport
	if os.Getenv("ENABLE_RAFT_LOG") != "true" {
		transport, err = raft.NewTCPTransport(raftBind, addr, 5, 5*time.Second, ioutil.Discard)
	} else {
		transport, err = raft.NewTCPTransport(raftBind, addr, 5, 5*time.Second, os.Stderr)
	}
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
	go s.leaderRoutine()
	index := s.raft.LastIndex()
	if index != 0 {
		return nil
	}
	go func() {
		for {
			// hardcoded for now
			err := s.joinCluster(name, 3)
			if err != nil {
				if err == ErrBootstrappedNodeFound {
					s.logger.Info("found an existing cluster, waiting for adoption")
					return
				}
				s.logger.Error("failed to join cluster, retrying", zap.Error(err))
			} else {
				return
			}
		}
	}()
	return nil
}

func (s *raftlayer) nodeStatus(node, serviceName string) string {
	serviceRPC := fmt.Sprintf("%s_rpc", serviceName)
	status := ""

	err := s.discovery.DialAddress(serviceRPC, node, func(c *grpc.ClientConn) error {
		client := pb.NewLayerClient(c)
		out, err := client.Status(context.TODO(), &pb.StatusInput{})
		if err != nil {
			return nil
		}
		status = out.Status
		return nil
	})
	if err != nil {
		return ""
	}
	return status
}

func (s *raftlayer) logRaftStatus(log *raft.Log) {
	s.logger.Debug("raft status", zap.Uint64("raft_last_index", s.raft.LastIndex()), zap.Uint64("raft_applied_index", s.raft.AppliedIndex()), zap.Uint64("raft_current_index", log.Index))
}
func (s *raftlayer) Health() string {
	if s.raft.AppliedIndex() == 0 {
		return "warning"
	}
	if (s.raft.Leader()) == "" {
		return "critical"
	}
	return "ok"
}

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

func (s *raftlayer) discoveredMembers() []string {
	members := []string{}
	discovered, err := s.discovery.Peers().EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
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

func (s *raftlayer) Status(ctx context.Context, input *pb.StatusInput) (*pb.StatusOutput, error) {
	return &pb.StatusOutput{
		Layer:  "raft",
		Status: s.status,
	}, nil
}
func (s *raftlayer) SendEvent(ctx context.Context, input *pb.SendEventInput) (*pb.SendEventOutput, error) {
	return &pb.SendEventOutput{}, s.ApplyEvent(input.Payload)
}
func (s *raftlayer) PrepareShutdown(ctx context.Context, input *pb.PrepareShutdownInput) (*pb.PrepareShutdownOutput, error) {
	return &pb.PrepareShutdownOutput{}, s.raft.RemoveServer(raft.ServerID(input.ID), input.Index, 500*time.Millisecond).Error()
}
func (s *raftlayer) IsLeader() bool {
	return s.raft.State() == raft.Leader
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
	err := s.raft.RemoveServer(raft.ServerID(id), 0, 2*time.Second).Error()
	if err != nil {
		s.logger.Error("failed to remove dead node", zap.Error(err))
		return
	}
	s.logger.Info("removed dead node", zap.Strings("raft_members", s.raftMembers()), zap.Strings("discovered_members", s.discoveredMembers()))
}
func (s *raftlayer) addMember(id, address string) {
	err := s.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(address), 0, 2*time.Second).Error()
	if err != nil {
		s.logger.Error("failed to add new node", zap.String("new_node", id), zap.Error(err))
		return
	}
	s.logger.Info("adopted new raft node", zap.String("new_node", id), zap.Strings("raft_members", s.raftMembers()), zap.Strings("discovered_members", s.discoveredMembers()), zap.Uint64("raft_last_index", s.raft.LastIndex()))
}
func (s *raftlayer) syncMembers() {
	members, err := s.discovery.Peers().EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
	if err != nil {
		s.logger.Error("failed to discover nodes", zap.Error(err))
		panic(err)
	}

	for _, member := range members {
		if member.Peer != s.id {
			if !s.isNodeAdopted(member.Peer) {
				s.addMember(member.Peer, member.NetworkAddress)
			}
		}
	}
	cProm := s.raft.GetConfiguration()
	err = cProm.Error()
	if err != nil {
		panic(err)
	}
	clusterConfig := cProm.Configuration()
	for _, server := range clusterConfig.Servers {
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
	ch := make(chan raft.Observation)
	s.raft.RegisterObserver(raft.NewObserver(ch, true, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.LeaderObservation)
		return ok
	}))
	s.discovery.Peers().On(peers.PeerCreated, func(member peers.Peer) {
		if !s.IsLeader() {
			return
		}
		s.syncMembers()
	})
	for range ch {
		s.logger.Info("leader changed", zap.String("raft_leader", string(s.raft.Leader())))
		leader := s.IsLeader()
		if s.status != raftStatusBootstrapped {
			s.logger.Info("raft cluster joined", zap.Uint64("raft_index", s.raft.LastIndex()),
				zap.Strings("discovered_members", s.discoveredMembers()),
			)
			s.status = raftStatusBootstrapped
		}
		if !leader {
			continue
		}
		s.logger.Info("raft cluster leadership acquired",
			zap.Strings("raft_members", s.raftMembers()), zap.Strings("discovered_members", s.discoveredMembers()),
		)
		s.syncMembers()
	}
}
func (s *raftlayer) Apply(log *raft.Log) interface{} {
	s.logRaftStatus(log)
	if log.Type == raft.LogCommand {
		err := s.state.Apply(log.Data)
		if err != nil {
			s.logger.Error("failed to apply raft event in FSM", zap.Error(err), zap.Uint64("raft_current_index", log.Index))
		}
	}
	return nil
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
	ctx := context.Background()
	ticker := time.NewTicker(3 * time.Second)
	s.logger.Info("shuting down raft layer")
	for {
		index := s.raft.LastIndex()
		if s.raft.VerifyLeader().Error() == nil {
			err := s.raft.LeadershipTransfer().Error()
			if err == nil {
				break
			}
			s.logger.Error("failed to transfert raft leadership", zap.Error(err))
		} else {
			err := s.DialLeader(func(client pb.LayerClient) error {
				_, err := client.PrepareShutdown(ctx, &pb.PrepareShutdownInput{
					ID:    s.id,
					Index: index,
				})
				return err
			})
			if err == nil {
				break
			}
			s.logger.Error("failed to leave raft cluster", zap.Error(err))
		}
		<-ticker.C
	}
	ticker.Stop()
	return s.raft.Shutdown().Error()
}
func (s *raftlayer) ApplyEvent(event []byte) error {
	if !s.IsLeader() {
		s.logger.Debug("forwarding event to leader")
		return s.DialLeader(func(client pb.LayerClient) error {
			_, err := client.SendEvent(context.TODO(), &pb.SendEventInput{Payload: event})
			if err != nil {
				s.logger.Error("failed to forward event to leader", zap.Error(err))
				return err
			}
			return nil
		})
	}
	promise := s.raft.Apply(event, 500*time.Millisecond)
	err := promise.Error()
	if err != nil {
		s.logger.Error("failed to apply raft event", zap.Error(err))
		return err
	}
	resp := promise.Response()
	if resp != nil {
		s.logger.Error("failed to apply raft event", zap.Error(resp.(error)))
		return resp.(error)
	}
	return nil
}
