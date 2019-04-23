package peers

import (
	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
)

var _ state.Backend = &memDBStore{}
var _ state.EntrySet = &PeerMetadataList{}

func (e *PeerMetadataList) Append(entry state.Entry) {
	sub := entry.(*Metadata)
	e.Peers = append(e.Peers, sub)
}
func (e *PeerMetadataList) Length() int {
	return len(e.Peers)
}
func (e *PeerMetadataList) New() state.EntrySet {
	return &PeerMetadataList{}
}
func (e *PeerMetadataList) AtIndex(idx int) state.Entry {
	return e.Peers[idx]
}
func (e *PeerMetadataList) Set(idx int, entry state.Entry) {
	sub := entry.(*Metadata)
	e.Peers[idx] = sub
}
func (e *PeerMetadataList) Range(f func(idx int, entry state.Entry)) {
	for idx, entry := range e.Peers {
		f(idx, entry)
	}
}

func (m *memDBStore) EntryByID(id string) (state.Entry, error) {
	var peer *Metadata
	err := m.read(func(tx *memdb.Txn) error {
		sess, err := m.first(tx, "id", id)
		if err != nil {
			return err
		}
		peer = sess
		return nil
	})
	return peer, err
}

func (m *memDBStore) InsertEntries(entries state.EntrySet) error {
	set := entries.(*PeerMetadataList)
	return m.insert(set.Peers)
}
func (m *memDBStore) InsertEntry(entry state.Entry) error {
	sub := entry.(*Metadata)
	return m.insert([]*Metadata{
		sub,
	})
}

func (m *memDBStore) DecodeSet(buf []byte) (state.EntrySet, error) {
	set := &PeerMetadataList{}
	return set, proto.Unmarshal(buf, set)
}
func (m memDBStore) Set() state.EntrySet {
	return &PeerMetadataList{}
}
func (m memDBStore) Dump() state.EntrySet {
	return m.DumpPeers()
}
func (m memDBStore) DumpPeers() *PeerMetadataList {
	peerList := PeerMetadataList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("peers", "id")
		if (err != nil && err != ErrPeerNotFound) || iterator == nil {
			return err
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*Metadata)
			peerList.Peers = append(peerList.Peers, sess)
		}
	})
	return &peerList
}

func (m *memDBStore) DeleteEntry(entry state.Entry) error {
	peer := entry.(*Metadata)
	return m.write(func(tx *memdb.Txn) error {
		err := tx.Delete("peers", peer)
		if err == nil {
			tx.Commit()
		}
		return err
	})
}
