package sessions

import (
	proto "github.com/golang/protobuf/proto"
	"github.com/weaveworks/mesh"
)

var _ mesh.GossipData = &SessionList{}
var _ mesh.Gossiper = &MemDBStore{}

func (set SessionList) Encode() [][]byte {
	payload, err := proto.Marshal(&set)
	if err != nil {
		return nil
	}
	return [][]byte{payload}
}

func (set SessionList) Merge(data mesh.GossipData) mesh.GossipData {
	remote := data.(*SessionList)
	idx := map[string]int{}
	set.ApplyIdx(func(i int, s *Session) {
		idx[s.ID] = i
	})
	delta := &SessionList{}
	remote.Apply(func(remote *Session) {
		local, ok := idx[remote.ID]
		if !ok {
			// Session not found in our store, add it to delta
			delta.Sessions = append(delta.Sessions, remote)
			set.Sessions = append(set.Sessions, remote)
		} else {
			if IsOutdated(set.Sessions[local], remote) {
				delta.Sessions = append(delta.Sessions, remote)
				set.Sessions[local] = remote
			}
		}
	})
	return delta
}

func (s *MemDBStore) Gossip() mesh.GossipData {
	return s.DumpState()
}
func (s *MemDBStore) merge(msg []byte) (mesh.GossipData, error) {
	set := &SessionList{}
	err := proto.Unmarshal(msg, set)
	if err != nil {
		return nil, err
	}
	delta := s.ComputeDelta(set)
	s.ApplyDelta(delta)
	return delta, nil
}
func (s *MemDBStore) OnGossip(msg []byte) (mesh.GossipData, error) {
	return s.merge(msg)
}
func (s *MemDBStore) OnGossipBroadcast(src mesh.PeerName, msg []byte) (mesh.GossipData, error) {
	return s.merge(msg)
}
func (s *MemDBStore) OnGossipUnicast(src mesh.PeerName, msg []byte) error {
	_, err := s.merge(msg)
	return err
}
