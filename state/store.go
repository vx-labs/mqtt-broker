package state

import (
	fmt "fmt"

	"github.com/golang/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/broker/cluster"
)

type Channel interface {
	Broadcast([]byte)
}

type Backend interface {
	EntryByID(id string) (Entry, error)
	InsertEntries(EntrySet) error
	InsertEntry(Entry) error
	DeleteEntry(Entry) error
	DecodeSet([]byte) (EntrySet, error)
	Set() EntrySet
	Dump() EntrySet
}
type Store struct {
	backend Backend
	channel Channel
}

func (s *Store) ApplyDelta(set EntrySet) error {
	return s.backend.InsertEntries(set)
}
func (s *Store) ComputeDelta(set EntrySet) EntrySet {
	delta := set.New()
	set.Range(func(_ int, remote Entry) {
		local, err := s.backend.EntryByID(remote.GetID())
		if err != nil {
			// Entry not found in our store, add it to delta
			delta.Append(remote)
		} else {
			if isEntryOutdated(local, remote) {
				delta.Append(remote)
			}
		}
	})
	return delta
}

func (s *Store) Merge(data []byte) error {
	set, err := s.backend.DecodeSet(data)
	if err != nil {
		return fmt.Errorf("failed to decode state data: %v", err)
	}
	return s.MergeEntries(set)
}
func (s *Store) MergeEntries(set EntrySet) error {
	delta := s.ComputeDelta(set)
	if delta.Length() == 0 {
		return nil
	}
	return s.ApplyDelta(delta)
}
func (s *Store) MarshalBinary() []byte {
	set := s.backend.Dump()
	payload, err := proto.Marshal(set)
	if err != nil {
		panic(err)
	}
	return payload
}
func (s *Store) Upsert(entry Entry) error {
	set := s.backend.Set()
	set.Append(entry)
	payload, err := proto.Marshal(set)
	if err != nil {
		return err
	}
	s.channel.Broadcast(payload)
	return s.backend.InsertEntry(entry)
}

func NewStore(name string, mesh cluster.Mesh, backend Backend) (*Store, error) {
	s := &Store{
		backend: backend,
	}
	channel, err := mesh.AddState(name, s)
	if err != nil {
		return nil, err
	}
	s.channel = channel
	go s.gCRunner()
	return s, nil
}
