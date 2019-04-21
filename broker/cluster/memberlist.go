package cluster

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/broker/cluster/ --go_out=plugins=grpc:. cluster.proto

import (
	"errors"
	"io/ioutil"
	"log"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/identity"
)

var (
	ErrStateKeyAlreadySet = errors.New("specified key is already taken")
	ErrNodeNotFound       = errors.New("specified node not found in mesh")
)

type memberlistMesh struct {
	mlist *memberlist.Memberlist

	mtx        sync.RWMutex
	states     map[string]State
	bcastQueue *memberlist.TransmitLimitedQueue
	meta       NodeMeta
}

func (m *memberlistMesh) ClusterSize() int {
	return m.mlist.NumMembers()
}

func (m *memberlistMesh) AddState(key string, state State) (Channel, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if _, ok := m.states[key]; ok {
		return nil, ErrStateKeyAlreadySet
	}
	m.states[key] = state
	userCh := &channel{
		bcast: m.bcastQueue,
		key:   key,
	}
	return userCh, nil
}
func (m *memberlistMesh) Join(hosts []string) error {
	_, err := m.mlist.Join(hosts)
	return err
}

func (m *memberlistMesh) NotifyMsg(b []byte) {
	var p Part
	if err := proto.Unmarshal(b, &p); err != nil {
		log.Printf("ERROR: failed to decode remote state: %v", err)
		return
	}
	s, ok := m.states[p.Key]
	if !ok {
		return
	}
	if err := s.Merge(p.Data); err != nil {
		log.Printf("ERROR: failed to merge remote %q state: %v", p.Key, err)
		return
	}
}
func (m *memberlistMesh) GetBroadcasts(overhead, limit int) [][]byte {
	return m.bcastQueue.GetBroadcasts(overhead, limit)
}
func (m *memberlistMesh) NodeMeta(limit int) []byte {
	payload, err := proto.Marshal(&m.meta)
	if err != nil || len(payload) > limit {
		log.Printf("WARN: not publishing node meta because limit of %d is exeeded (%d)", limit, len(payload))
		return []byte{}
	}
	return payload
}
func (m *memberlistMesh) LocalState(join bool) []byte {
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
func (m *memberlistMesh) MergeRemoteState(buf []byte, join bool) {
	var fs FullState
	if err := proto.Unmarshal(buf, &fs); err != nil {
		log.Printf("ERROR: failed to decode remote full state: %v", err)
		return
	}
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	for _, p := range fs.Parts {
		s, ok := m.states[p.Key]
		if !ok {
			continue
		}
		if err := s.Merge(p.Data); err != nil {
			log.Printf("ERROR: failed to merge remote full %q state: %v", p.Key, err)
			return
		}
	}
}

func (m *memberlistMesh) MemberRPCAddress(id string) (string, error) {
	for _, node := range m.mlist.Members() {
		if node.Name == id {
			var meta NodeMeta
			err := proto.Unmarshal(node.Meta, &meta)
			if err != nil {
				break
			}
			return meta.RPCAddr, nil
		}
	}
	return "", ErrNodeNotFound
}

func (m *memberlistMesh) ID() string {
	return m.mlist.LocalNode().Name
}

func MemberlistMesh(id identity.Identity, eventHandler memberlist.EventDelegate, meta NodeMeta) Mesh {
	self := &memberlistMesh{
		states: map[string]State{},
		meta:   meta,
	}
	config := memberlist.DefaultLANConfig()
	config.AdvertiseAddr = id.Public().Host()
	config.AdvertisePort = id.Public().Port()
	config.BindPort = id.Private().Port()
	config.Name = id.ID()
	config.Delegate = self
	config.Events = eventHandler
	config.LogOutput = ioutil.Discard
	list, err := memberlist.Create(config)
	if err != nil {
		log.Fatal("failed to create memberlist: " + err.Error())
	}
	self.mlist = list
	self.bcastQueue = &memberlist.TransmitLimitedQueue{
		NumNodes:       self.ClusterSize,
		RetransmitMult: 3,
	}
	return self
}
