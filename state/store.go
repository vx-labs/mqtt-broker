package state

import "github.com/weaveworks/mesh"

type Backend interface {
	EntryByID(id string) (Entry, error)
	InsertEntries(EntrySet) error
	InsertEntry(Entry) error
	DeleteEntry(Entry) error
	DecodeSet([]byte) (EntrySet, error)
	Set() EntrySet
	Dump() EntrySet
}
type Router interface {
	NewGossip(channel string, gossiper mesh.Gossiper) (mesh.Gossip, error)
}

var _ mesh.Gossiper = &Store{}

type Store struct {
	backend Backend
	gossip  mesh.Gossip
}

func (s *Store) ApplyDelta(set EntrySet) error {
	return s.backend.InsertEntries(set)
}
func (s *Store) ComputeDelta(set EntrySet) EntrySet {
	delta := set.New()
	set.Range(func(_ int, remote Entry) {
		local, err := s.backend.EntryByID(remote.GetID())
		if err != nil {
			// Session not found in our store, add it to delta
			delta.Append(remote)
		} else {
			if isEntryOutdated(local, remote) {
				delta.Append(remote)
			}
		}
	})
	return delta
}

func (s *Store) Gossip() mesh.GossipData {
	return &Dataset{
		backend: s.backend.Dump(),
	}
}
func (s *Store) merge(msg []byte) (mesh.GossipData, error) {
	set, err := s.backend.DecodeSet(msg)
	if err != nil {
		return nil, err
	}
	delta := s.ComputeDelta(set)
	return &Dataset{
		backend: delta,
	}, s.ApplyDelta(delta)
}
func (s *Store) OnGossip(msg []byte) (mesh.GossipData, error) {
	return s.merge(msg)
}
func (s *Store) OnGossipBroadcast(src mesh.PeerName, msg []byte) (mesh.GossipData, error) {
	return s.merge(msg)
}
func (s *Store) OnGossipUnicast(src mesh.PeerName, msg []byte) error {
	_, err := s.merge(msg)
	return err
}

func (s *Store) Upsert(entry Entry) error {
	set := s.backend.Set()
	set.Append(entry)
	s.gossip.GossipBroadcast(&Dataset{backend: set})
	return s.backend.InsertEntry(entry)
}

func NewStore(channel string, backend Backend, router Router) (*Store, error) {
	s := &Store{
		backend: backend,
	}
	gossip, err := router.NewGossip(channel, s)
	if err != nil {
		return nil, err
	}
	s.gossip = gossip
	return s, nil
}
