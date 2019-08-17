package layer

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

func NewRaftLayer(logger *zap.Logger, userConfig config.Config, discovery DiscoveryProvider) (*raftlayer, error) {
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
	transport, err := raft.NewTCPTransport(raftBind, addr, 3, 10*time.Second, ioutil.Discard)
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
	go s.leaderRoutine()
	// hardcoded for now
	for {
		err := s.joinCluster(name, 3)
		if err != nil {
			if err == ErrBootstrappedNodeFound {
				s.logger.Info("found an existing cluster, waiting for adoption")
				return nil
			}
			s.logger.Error("failed to join cluster, retrying", zap.Error(err))
		} else {
			return nil
		}
	}
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

func (s *raftlayer) waitForToken(token string, expectNodeCount int) error {
	ticker := time.NewTicker(5 * time.Second)
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
			s.logger.Info("waiting for other members to complete step 2")
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
			s.logger.Info("found other members", zap.Int("member_count", len(bootstrappingMembers)))
			members = bootstrappingMembers
			break
		}
		s.logger.Info("waiting for other members to be discovered")
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

func (s *raftlayer) Status(ctx context.Context, input *pb.StatusInput) (*pb.StatusOutput, error) {
	return &pb.StatusOutput{
		Layer:  "raft",
		Status: s.status,
	}, nil
}
func (s *raftlayer) SendEvent(ctx context.Context, input *pb.SendEventInput) (*pb.SendEventOutput, error) {
	return &pb.SendEventOutput{}, s.ApplyEvent(input.Payload)
}
func (s *raftlayer) isLeader() bool {
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
	err := s.raft.RemoveServer(raft.ServerID(id), 0, 100*time.Millisecond).Error()
	if err != nil {
		s.logger.Error("failed to remove dead node", zap.Error(err))
		return
	}
	s.logger.Info("removed dead node")
}
func (s *raftlayer) addMember(id, address string) {
	err := s.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(address), 0, 100*time.Millisecond).Error()
	if err != nil {
		s.logger.Error("failed to add new node", zap.String("new_node", id), zap.Error(err))
		return
	}
	s.logger.Info("adopted new raft node", zap.String("new_node", id))
}
func (s *raftlayer) syncMembers() {
	members, err := s.discovery.Peers().EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
	if err != nil {
		s.logger.Error("failed to discover nodes", zap.Error(err))
		panic(err)
	}

	for _, member := range members {
		if member.Peer != s.id {
			s.addMember(member.Peer, member.NetworkAddress)
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
	s.discovery.Peers().On(peers.PeerCreated, func(member peers.Peer) {
		if !s.isLeader() {
			return
		}
		if s.isNodeAdopted(member.ID) {
			return
		}
		for _, service := range member.HostedServices {
			if service.ID == fmt.Sprintf("%s_cluster", s.name) {
				s.addMember(service.Peer, service.NetworkAddress)
			}
		}
	})
	s.discovery.Peers().On(peers.PeerDeleted, func(member peers.Peer) {
		if !s.isLeader() {
			return
		}
		if !s.isNodeAdopted(member.ID) {
			return
		}
		for _, service := range member.HostedServices {
			if service.ID == fmt.Sprintf("%s_cluster", s.name) {
				s.removeMember(string(service.Peer))
			}
		}
	})
	for leader := range s.raft.LeaderCh() {
		if s.status != raftStatusBootstrapped {
			s.logger.Info("raft cluster joined")
			s.status = raftStatusBootstrapped
		}
		if !leader {
			s.logger.Info("raft cluster leadership lost")
			continue
		}
		s.logger.Info("raft cluster leadership acquired")
		s.syncMembers()
	}
}
func (s *raftlayer) Apply(log *raft.Log) interface{} {
	return s.state.Apply(log.Data, s.isLeader())
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
	if !s.isLeader() {
		s.logger.Debug("forwarding event to leader")
		leader := string(s.raft.Leader())
		serviceRPC := fmt.Sprintf("%s_rpc", s.name)
		members, err := s.discovery.Peers().EndpointsByService(fmt.Sprintf("%s_cluster", s.name))
		if err != nil {
			s.logger.Error("failed to discover nodes", zap.Error(err))
			return err
		}
		memberAddresses := make([]string, len(members))
		for idx, member := range members {
			if member.NetworkAddress == leader {
				memberAddresses[idx] = member.NetworkAddress
				return s.discovery.DialAddress(serviceRPC, member.Peer, func(c *grpc.ClientConn) error {
					client := pb.NewLayerClient(c)
					_, err := client.SendEvent(context.TODO(), &pb.SendEventInput{Payload: event})
					if err != nil {
						return err
					}
					return nil
				})
			}
		}
		s.logger.Error("failed to find leader addess", zap.String("leader_address", leader), zap.Strings("members_addresses", memberAddresses))
		return errors.New("failed to find leader address")
	}
	promise := s.raft.Apply(event, 300*time.Millisecond)
	err := promise.Error()
	if err != nil {
		return err
	}
	resp := promise.Response()
	if resp != nil {
		return resp.(error)
	}
	return s.raft.Barrier(300 * time.Millisecond).Error()
}
