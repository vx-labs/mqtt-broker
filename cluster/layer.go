package cluster

import (
	fmt "fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
)

type Config struct {
	ID            string
	AdvertiseAddr string
	AdvertisePort int
	BindPort      int
	onNodeJoin    func(id string, meta pb.NodeMeta)
	onNodeLeave   func(id string, meta pb.NodeMeta)
}

type layer struct {
	id          string
	name        string
	mlist       *memberlist.Memberlist
	logger      *zap.Logger
	mtx         sync.RWMutex
	states      map[string]types.State
	bcastQueue  *memberlist.TransmitLimitedQueue
	meta        pb.NodeMeta
	onNodeJoin  func(id string, meta pb.NodeMeta)
	onNodeLeave func(id string, meta pb.NodeMeta)
}

func (m *layer) Members() []*memberlist.Node {
	return m.mlist.Members()
}

func (m *layer) AddState(key string, state types.State) (types.Channel, error) {
	//log.Printf("INFO: service/%s: registering %s state", m.name, key)
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
func (m *layer) NotifyMsg(b []byte) {
	var p pb.Part
	if err := proto.Unmarshal(b, &p); err != nil {
		m.logger.Error("failed to decode remote state", zap.String("node_id", m.id), zap.Error(err))
		return
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()
	s, ok := m.states[p.Key]
	if !ok {
		m.states[p.Key] = &cachedState{
			data: p.Data,
		}
		return
	}
	if err := s.Merge(p.Data, false); err != nil {
		m.logger.Error("failed to merge remote state", zap.String("node_id", m.id), zap.Error(err))
		return
	}
}
func (m *layer) GetBroadcasts(overhead, limit int) [][]byte {
	return m.bcastQueue.GetBroadcasts(overhead, limit)
}
func (m *layer) NodeMeta(limit int) []byte {
	payload, err := proto.Marshal(&m.meta)
	if err != nil || len(payload) > limit {
		m.logger.Warn("not publishing node meta because limit is exceeded", zap.String("node_id", m.id))
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
		m.logger.Error("failed to marshal full state", zap.String("node_id", m.id), zap.Error(err))
		return nil
	}
	return payload
}
func (m *layer) MergeRemoteState(buf []byte, join bool) {
	var fs pb.FullState
	if err := proto.Unmarshal(buf, &fs); err != nil {
		m.logger.Error("failed to decode remote state", zap.String("node_id", m.id), zap.Error(err))
		return
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, p := range fs.Parts {
		s, ok := m.states[p.Key]
		if !ok {
			m.states[p.Key] = &cachedState{
				data: p.Data,
			}
			continue
		}
		//now := time.Now()
		if err := s.Merge(p.Data, join); err != nil {
			m.logger.Error("failed to merge remote state", zap.String("node_id", m.id), zap.Error(err))
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
	for _, host := range newHosts {
		for _, curHost := range curHosts {
			found := false
			if curHost.Address() == host {
				found = true
				break
			}
			if !found {
				hosts = append(hosts, host)
			}
		}
	}
	if len(hosts) == 0 {
		return nil
	}
	m.logger.Info("joining cluster", zap.String("node_id", m.id), zap.Strings("nodes", hosts))
	count, err := m.mlist.Join(hosts)
	if err != nil {
		if count == 0 {
			m.logger.Error("failed to join cluster", zap.String("node_id", m.id), zap.Error(err))
			return err
		}
		m.logger.Warn("failed to join some member of cluster", zap.String("node_id", m.id), zap.Error(err))
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
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			err := self.Join(addresses)
			if err == nil {
				return
			}
			<-ticker.C
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

func NewLayer(name string, logger *zap.Logger, userConfig Config, meta pb.NodeMeta) Layer {
	self := &layer{
		id:          userConfig.ID,
		name:        name,
		states:      map[string]types.State{},
		meta:        meta,
		onNodeJoin:  userConfig.onNodeJoin,
		onNodeLeave: userConfig.onNodeLeave,
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
