package cluster

import (
	fmt "fmt"
	"log"
	"sync"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/hashicorp/memberlist"
)

type Config struct {
	ID            string
	AdvertiseAddr string
	AdvertisePort int
	BindPort      int
	onNodeJoin    func(id string, meta NodeMeta)
	onNodeLeave   func(id string, meta NodeMeta)
}

type layer struct {
	id    string
	name  string
	mlist *memberlist.Memberlist

	mtx         sync.RWMutex
	states      map[string]State
	bcastQueue  *memberlist.TransmitLimitedQueue
	meta        NodeMeta
	onNodeJoin  func(id string, meta NodeMeta)
	onNodeLeave func(id string, meta NodeMeta)
}

func (m *layer) Members() []*memberlist.Node {
	return m.mlist.Members()
}

func (m *layer) AddState(key string, state State) (Channel, error) {
	log.Printf("INFO: service/%s: registering %s state", m.name, key)
	m.mtx.Lock()
	defer m.mtx.Unlock()
	old, ok := m.states[key]
	m.states[key] = state
	if ok {
		log.Printf("INFO: service/%s: merging cached state into new state %q", m.name, key)
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
	var p Part
	if err := proto.Unmarshal(b, &p); err != nil {
		log.Printf("ERROR: failed to decode remote state: %v", err)
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
		log.Printf("ERROR: failed to merge remote %q state: %v", p.Key, err)
		return
	}
}
func (m *layer) GetBroadcasts(overhead, limit int) [][]byte {
	return m.bcastQueue.GetBroadcasts(overhead, limit)
}
func (m *layer) NodeMeta(limit int) []byte {
	payload, err := proto.Marshal(&m.meta)
	if err != nil || len(payload) > limit {
		log.Printf("WARN: not publishing node meta because limit of %d is exeeded (%d)", limit, len(payload))
		return []byte{}
	}
	return payload
}
func (m *layer) LocalState(join bool) []byte {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	dump := &FullState{
		Parts: make([]*Part, 0, len(m.states)),
	}
	for key, state := range m.states {
		dump.Parts = append(dump.Parts, &Part{Key: key, Data: state.MarshalBinary()})
	}
	payload, err := proto.Marshal(dump)
	if err != nil {
		log.Printf("ERROR: failed to marshal full state: %v", err)
		return nil
	}
	return payload
}
func (m *layer) MergeRemoteState(buf []byte, join bool) {
	var fs FullState
	if err := proto.Unmarshal(buf, &fs); err != nil {
		log.Printf("ERROR: failed to decode remote full state: %v", err)
		return
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, p := range fs.Parts {
		s, ok := m.states[p.Key]
		if !ok {
			log.Printf("WARN: caching state for key %s", p.Key)
			m.states[p.Key] = &cachedState{
				data: p.Data,
			}
			continue
		}
		now := time.Now()
		if err := s.Merge(p.Data, join); err != nil {
			log.Printf("ERROR: failed to merge remote full %q state: %v", p.Key, err)
			return
		}
		log.Printf("DEBUG: %s merge done (%s elapsed)", p.Key, time.Now().Sub(now).String())
	}
}

func (m *layer) Join(hosts []string) {
	if len(hosts) == 0 {
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		retry := 5
		for retry > 0 {
			log.Printf("INFO: service/%s: joining hosts %v", m.name, hosts)
			_, err := m.mlist.Join(hosts)
			if err == nil {
				return
			}
			log.Printf("WARN: service/%s: failed to join the provided node list: %v", m.name, err)
			<-ticker.C
			retry--
		}
	}()
}
func (self *layer) DiscoverPeers(discovery PeerStore) {
	peers, _ := discovery.EndpointsByService(fmt.Sprintf("%s_cluster", self.name))
	addresses := []string{}
	for _, peer := range peers {
		if !self.isNodeKnown(peer.Peer) {
			addresses = append(addresses, peer.NetworkAddress)
		}
	}
	if len(addresses) > 0 {
		self.Join(addresses)
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

func NewLayer(name string, userConfig Config, meta NodeMeta) Layer {
	self := &layer{
		id:          userConfig.ID,
		name:        name,
		states:      map[string]State{},
		meta:        meta,
		onNodeJoin:  userConfig.onNodeJoin,
		onNodeLeave: userConfig.onNodeLeave,
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
	//config.LogOutput = ioutil.Discard
	list, err := memberlist.Create(config)
	if err != nil {
		panic(err)
	}
	self.mlist = list
	return self
}