package topics

import (
	proto "github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/weaveworks/mesh"
)

var _ mesh.GossipData = &RetainedMessageList{}
var _ mesh.Gossiper = &memDBStore{}

func (set RetainedMessageList) Encode() [][]byte {
	payload, err := proto.Marshal(&set)
	if err != nil {
		return nil
	}
	return [][]byte{payload}
}

func (set RetainedMessageList) Merge(data mesh.GossipData) mesh.GossipData {
	remote := data.(*RetainedMessageList)
	idx := map[string]int{}
	set.ApplyIdx(func(i int, s *RetainedMessage) {
		idx[s.Id] = i
	})
	delta := &RetainedMessageList{}
	remote.Apply(func(remote *RetainedMessage) {
		local, ok := idx[remote.Id]
		if !ok {
			// RetainedMessage not found in our store, add it to delta
			delta.RetainedMessages = append(delta.RetainedMessages, remote)
			set.RetainedMessages = append(set.RetainedMessages, remote)
		} else {
			if IsOutdated(set.RetainedMessages[local], remote) {
				delta.RetainedMessages = append(delta.RetainedMessages, remote)
				set.RetainedMessages[local] = remote
			}
		}
	})
	return delta
}

func (s *memDBStore) DumpState() *RetainedMessageList {
	sessionList := RetainedMessageList{}
	s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("messages", "id")
		if err != nil || iterator == nil {
			return ErrRetainedMessageNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*RetainedMessage)
			sessionList.RetainedMessages = append(sessionList.RetainedMessages, sess)
		}
	})
	return &sessionList
}
func (s *memDBStore) ComputeDelta(set *RetainedMessageList) *RetainedMessageList {
	delta := RetainedMessageList{}
	set.Apply(func(remote *RetainedMessage) {
		local, err := s.ByID(remote.Id)
		if err != nil {
			// Session not found in our store, add it to delta
			delta.RetainedMessages = append(delta.RetainedMessages, remote)
		} else {
			if IsOutdated(local, remote) {
				delta.RetainedMessages = append(delta.RetainedMessages, remote)
			}
		}
	})
	return &delta
}
func (s *memDBStore) ApplyDelta(set *RetainedMessageList) {
	set.Apply(func(remote *RetainedMessage) {
		s.insert(remote)
	})
}

func (s *memDBStore) Gossip() mesh.GossipData {
	return s.DumpState()
}
func (s *memDBStore) merge(msg []byte) (mesh.GossipData, error) {
	set := &RetainedMessageList{}
	err := proto.Unmarshal(msg, set)
	if err != nil {
		return nil, err
	}
	delta := s.ComputeDelta(set)
	s.ApplyDelta(delta)
	return delta, nil
}
func (s *memDBStore) OnGossip(msg []byte) (mesh.GossipData, error) {
	return s.merge(msg)
}
func (s *memDBStore) OnGossipBroadcast(src mesh.PeerName, msg []byte) (mesh.GossipData, error) {
	return s.merge(msg)
}
func (s *memDBStore) OnGossipUnicast(src mesh.PeerName, msg []byte) error {
	_, err := s.merge(msg)
	return err
}
