package layer

import (
	"context"
	fmt "fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
)

type layer struct {
	id          string
	name        string
	mlist       *memberlist.Memberlist
	logger      *zap.Logger
	mtx         sync.RWMutex
	states      map[string]types.GossipState
	bcastQueue  *memberlist.TransmitLimitedQueue
	meta        pb.NodeMeta
	onNodeJoin  func(id string, meta pb.NodeMeta)
	onNodeLeave func(id string, meta pb.NodeMeta)
}

func (s *layer) Status(ctx context.Context, input *pb.StatusInput) (*pb.StatusOutput, error) {
	return &pb.StatusOutput{
		Layer:  "gossip",
		Status: "ok",
	}, nil
}
func (m *layer) Members() []*memberlist.Node {
	return m.mlist.Members()
}

func (m *layer) OnNodeJoin(f func(id string, meta pb.NodeMeta)) {
	m.onNodeJoin = f
}
func (m *layer) OnNodeLeave(f func(id string, meta pb.NodeMeta)) {
	m.onNodeLeave = f
}

func (m *layer) AddState(key string, state types.GossipState) (types.Channel, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	old, ok := m.states[key]
	m.states[key] = state
	if ok {
		err := state.Merge(old.MarshalBinary(), true)
		if err != nil {
			return nil, err
		}
	}
	userCh := &channel{
		bcast: m.bcastQueue,
		key:   key,
	}
	return userCh, nil
}
func (s *layer) SendEvent(ctx context.Context, input *pb.SendEventInput) (*pb.SendEventOutput, error) {
	s.NotifyMsg(input.Payload)
	return &pb.SendEventOutput{}, nil
}
func (m *layer) NotifyMsg(b []byte) {
	var p pb.Part
	if err := proto.Unmarshal(b, &p); err != nil {
		m.logger.Error("failed to decode remote state", zap.Error(err))
		return
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()
	s, ok := m.states[p.Key]
	if !ok {
		return
	}
	if err := s.Merge(p.Data, false); err != nil {
		m.logger.Error("failed to merge remote state", zap.Error(err))
		return
	}
}
func (s *layer) Health() string {
	if s.mlist.NumMembers() == 1 {
		return "warning"
	}
	return "ok"
}
func (m *layer) GetBroadcasts(overhead, limit int) [][]byte {
	return m.bcastQueue.GetBroadcasts(overhead, limit)
}
func (m *layer) NodeMeta(limit int) []byte {
	payload, err := proto.Marshal(&m.meta)
	if err != nil || len(payload) > limit {
		m.logger.Warn("not publishing node meta because limit is exceeded")
		return []byte{}
	}
	return payload
}
func (m *layer) LocalState(join bool) []byte {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	dump := &pb.FullState{
		Parts: make([]*pb.Part, 0, len(m.states)),
	}
	for key, state := range m.states {
		dump.Parts = append(dump.Parts, &pb.Part{Key: key, Data: state.MarshalBinary()})
	}
	payload, err := proto.Marshal(dump)
	if err != nil {
		m.logger.Error("failed to marshal full state", zap.Error(err))
		return nil
	}
	return payload
}
func (m *layer) MergeRemoteState(buf []byte, join bool) {
	var fs pb.FullState
	if err := proto.Unmarshal(buf, &fs); err != nil {
		m.logger.Error("failed to decode remote state", zap.Error(err))
		return
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, p := range fs.Parts {
		s, ok := m.states[p.Key]
		if !ok {
			continue
		}
		//now := time.Now()
		if err := s.Merge(p.Data, join); err != nil {
			m.logger.Error("failed to merge remote state", zap.Error(err))
			return
		}
	}
}

func (m *layer) Join(newHosts []string) error {
	if len(newHosts) == 0 {
		return nil
	}
	hosts := []string{}
	curHosts := m.mlist.Members()
	for idx := range newHosts {
		host := newHosts[idx]
		found := false
		for idx := range curHosts {
			curHost := curHosts[idx]
			if curHost.Address() == host {
				found = true
				break
			}
		}
		if !found {
			hosts = append(hosts, host)
		}
	}
	if len(hosts) == 0 {
		return nil
	}
	if m.mlist.NumMembers() == 1 {
		m.logger.Info("joining cluster", zap.Strings("nodes", hosts), zap.Strings("provided_nodes", newHosts))
	}
	count, err := m.mlist.Join(hosts)
	if err != nil {
		if count == 0 {
			m.logger.Error("failed to join cluster", zap.Error(err))
			return err
		}
		m.logger.Warn("failed to join some member of cluster", zap.Error(err))
	}
	return nil
}
func (self *layer) DiscoverPeers(discovery peers.PeerStore) {
	peers, _ := discovery.EndpointsByService(fmt.Sprintf("%s_cluster", self.name))
	addresses := []string{}
	for _, peer := range peers {
		if !self.isNodeKnown(peer.Peer) {
			addresses = append(addresses, peer.NetworkAddress)
		}
	}
	if len(addresses) > 0 {
		err := self.Join(addresses)
		if err == nil {
			return
		}
	}
}
func (self *layer) Leave() {
	self.mlist.Leave(5 * time.Second)
}
func (self *layer) isNodeKnown(id string) bool {
	members := self.mlist.Members()
	for _, member := range members {
		if member.Name == id {
			return true
		}
	}
	return false
}
func (self *layer) numMembers() int {
	if self.mlist == nil {
		return 1
	}
	return self.mlist.NumMembers()
}

func NewGossipLayer(name string, logger *zap.Logger, userConfig config.Config, meta pb.NodeMeta) *layer {
	self := &layer{
		id:          userConfig.ID,
		name:        name,
		states:      map[string]types.GossipState{},
		meta:        meta,
		onNodeJoin:  userConfig.OnNodeJoin,
		onNodeLeave: userConfig.OnNodeLeave,
		logger:      logger,
	}

	self.bcastQueue = &memberlist.TransmitLimitedQueue{
		NumNodes:       self.numMembers,
		RetransmitMult: 3,
	}

	config := memberlist.DefaultLANConfig()
	config.AdvertiseAddr = userConfig.AdvertiseAddr
	config.AdvertisePort = userConfig.AdvertisePort
	config.BindPort = userConfig.BindPort
	config.Name = userConfig.ID
	config.Delegate = self
	config.Events = self
	if os.Getenv("ENABLE_MEMBERLIST_LOG") != "true" {
		config.LogOutput = ioutil.Discard
	}
	list, err := memberlist.Create(config)
	if err != nil {
		panic(err)
	}
	self.mlist = list
	return self
}
