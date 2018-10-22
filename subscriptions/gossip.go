package subscriptions

import (
	proto "github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/weaveworks/mesh"
)

var _ mesh.GossipData = &SubscriptionList{}
var _ mesh.Gossiper = &memDBStore{}

func (set SubscriptionList) Encode() [][]byte {
	payload, err := proto.Marshal(&set)
	if err != nil {
		return nil
	}
	return [][]byte{payload}
}

func (set SubscriptionList) Merge(data mesh.GossipData) mesh.GossipData {
	remote := data.(*SubscriptionList)
	idx := map[string]int{}
	set.ApplyIdx(func(i int, s *Subscription) {
		idx[s.ID] = i
	})
	delta := &SubscriptionList{}
	remote.Apply(func(remote *Subscription) {
		local, ok := idx[remote.ID]
		if !ok {
			// Subscription not found in our store, add it to delta
			delta.Subscriptions = append(delta.Subscriptions, remote)
			set.Subscriptions = append(set.Subscriptions, remote)
		} else {
			if IsOutdated(set.Subscriptions[local], remote) {
				delta.Subscriptions = append(delta.Subscriptions, remote)
				set.Subscriptions[local] = remote
			}
		}
	})
	return delta
}

func (s *memDBStore) DumpState() *SubscriptionList {
	sessionList := SubscriptionList{}
	s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("subscriptions", "id")
		if err != nil || iterator == nil {
			return ErrSubscriptionNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*Subscription)
			sessionList.Subscriptions = append(sessionList.Subscriptions, sess)
		}
	})
	return &sessionList
}
func (s *memDBStore) ComputeDelta(set *SubscriptionList) *SubscriptionList {
	delta := SubscriptionList{}
	set.Apply(func(remote *Subscription) {
		local, err := s.ByID(remote.ID)
		if err != nil {
			// Session not found in our store, add it to delta
			delta.Subscriptions = append(delta.Subscriptions, remote)
		} else {
			if IsOutdated(local, remote) {
				delta.Subscriptions = append(delta.Subscriptions, remote)
			}
		}
	})
	return &delta
}
func (s *memDBStore) ApplyDelta(set *SubscriptionList) {
	set.Apply(func(remote *Subscription) {
		s.insert(remote)
	})
}

func (s *memDBStore) Gossip() mesh.GossipData {
	return s.DumpState()
}
func (s *memDBStore) merge(msg []byte) (mesh.GossipData, error) {
	set := &SubscriptionList{}
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
