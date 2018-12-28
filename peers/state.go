package peers

import (
	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
)

var _ state.Backend = &memDBStore{}
var _ state.EntrySet = &PeerList{}

func (e *PeerList) Append(entry state.Entry) {
	sub := entry.(*Peer)
	e.Peers = append(e.Peers, sub)
}
func (e *PeerList) Length() int {
	return len(e.Peers)
}
func (e *PeerList) New() state.EntrySet {
	return &PeerList{}
}
func (e *PeerList) AtIndex(idx int) state.Entry {
	return e.Peers[idx]
}
func (e *PeerList) Set(idx int, entry state.Entry) {
	sub := entry.(*Peer)
	e.Peers[idx] = sub
}
func (e *PeerList) Range(f func(idx int, entry state.Entry)) {
	for idx, entry := range e.Peers {
		f(idx, entry)
	}
}

func (m *memDBStore) EntryByID(id string) (state.Entry, error) {
	var peer *Peer
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
	set := entries.(*PeerList)
	return m.insert(set.Peers)
}
func (m *memDBStore) InsertEntry(entry state.Entry) error {
	sub := entry.(*Peer)
	return m.insert([]*Peer{
		sub,
	})
}

func (m *memDBStore) DecodeSet(buf []byte) (state.EntrySet, error) {
	set := &PeerList{}
	return set, proto.Unmarshal(buf, set)
}
func (m memDBStore) Set() state.EntrySet {
	return &PeerList{}
}
func (m memDBStore) Dump() state.EntrySet {
	peerList := PeerList{}
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
			sess := payload.(*Peer)
			peerList.Peers = append(peerList.Peers, sess)
		}
	})
	return &peerList
}

func (m *memDBStore) DeleteEntry(entry state.Entry) error {
	peer := entry.(*Peer)
	return m.write(func(tx *memdb.Txn) error {
		err := tx.Delete("peers", peer)
		if err == nil {
			tx.Commit()
		}
		return err
	})
}
