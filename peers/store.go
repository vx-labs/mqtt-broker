package peers

import (
	"errors"
	"time"

	"github.com/vx-labs/mqtt-broker/events"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
	"github.com/weaveworks/mesh"
)

const (
	defaultTenant = "_default"
)
const (
	PeerCreated string = "peer_created"
	PeerDeleted string = "peer_deleted"
)

var (
	ErrPeerNotFound = errors.New("peer not found")
)

var now = func() int64 {
	return time.Now().UnixNano()
}

type PeerStore interface {
	ByID(id string) (*Peer, error)
	ByMeshID(id uint64) (*Peer, error)
	All() (PeerList, error)
	Exists(id string) bool
	Upsert(p *Peer) error
	Delete(id string) error
	On(event string, handler func(*Peer)) func()
}

type memDBStore struct {
	db     *memdb.MemDB
	state  *state.Store
	events *events.Bus
}

type Router interface {
	NewGossip(channel string, gossiper mesh.Gossiper) (mesh.Gossip, error)
}

func NewPeerStore(router Router) (PeerStore, error) {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"peers": {
				Name: "peers",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name: "id",
						Indexer: &memdb.StringFieldIndex{
							Field: "ID",
						},
						Unique:       true,
						AllowMissing: false,
					},
					"meshID": &memdb.IndexSchema{
						Name:         "meshID",
						AllowMissing: false,
						Unique:       true,
						Indexer:      &memdb.UintFieldIndex{Field: "MeshID"},
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	s := &memDBStore{
		db:     db,
		events: events.NewEventBus(),
	}
	state, err := state.NewStore("mqtt-peers", s, router)
	if err != nil {
		return nil, err
	}
	s.state = state
	return s, nil
}
func (s *memDBStore) all() *PeerList {
	peerList := PeerList{}
	s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("peers", "id")
		if err != nil || iterator == nil {
			return ErrPeerNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			p := payload.(*Peer)
			peerList.Peers = append(peerList.Peers, p)
		}
	})
	return &peerList
}

func (s *memDBStore) Exists(id string) bool {
	_, err := s.ByID(id)
	return err == nil
}
func (s *memDBStore) ByID(id string) (*Peer, error) {
	var peer *Peer
	return peer, s.read(func(tx *memdb.Txn) error {
		p, err := s.first(tx, "id", id)
		if err != nil {
			return err
		}
		if p.IsRemoved() {
			return ErrPeerNotFound
		}
		peer = p
		return nil
	})
}
func (s *memDBStore) All() (PeerList, error) {
	return s.all().Filter(func(s *Peer) bool {
		return s.IsAdded()
	}), nil
}

func (s *memDBStore) ByMeshID(id uint64) (*Peer, error) {
	var peer *Peer
	err := s.read(func(tx *memdb.Txn) error {
		data, err := tx.First("peers", "meshID", id)
		if err != nil || data == nil {
			return ErrPeerNotFound
		}
		p := data.(*Peer)
		peer = p
		return nil
	})
	return peer, err
}

func (s *memDBStore) Upsert(p *Peer) error {
	p.LastAdded = now()
	return s.state.Upsert(p)
}
func (s *memDBStore) insert(peers []*Peer) error {
	return s.write(func(tx *memdb.Txn) error {
		for _, p := range peers {
			if p.IsAdded() {
				s.events.Emit(events.Event{
					Entry: p,
					Key:   PeerCreated,
				})
				s.events.Emit(events.Event{
					Entry: p,
					Key:   PeerCreated + "/" + p.ID,
				})
			} else if p.IsRemoved() {
				s.events.Emit(events.Event{
					Entry: p,
					Key:   PeerDeleted,
				})
				s.events.Emit(events.Event{
					Entry: p,
					Key:   PeerDeleted + "/" + p.ID,
				})
			}
			err := tx.Insert("peers", p)
			if err != nil {
				return err
			}
		}
		tx.Commit()
		return nil
	})
}
func (s *memDBStore) Delete(id string) error {
	p, err := s.ByID(id)
	if err != nil {
		return nil
	}
	p.LastDeleted = now()
	return s.state.Upsert(p)
}

func (s *memDBStore) read(statement func(tx *memdb.Txn) error) error {
	tx := s.db.Txn(false)
	return s.run(tx, statement)
}
func (s *memDBStore) write(statement func(tx *memdb.Txn) error) error {
	tx := s.db.Txn(true)
	return s.run(tx, statement)
}
func (s *memDBStore) run(tx *memdb.Txn, statement func(tx *memdb.Txn) error) error {
	defer tx.Abort()
	err := statement(tx)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (s *memDBStore) first(tx *memdb.Txn, idx, id string) (*Peer, error) {
	data, err := tx.First("peers", idx, id)
	if err != nil || data == nil {
		return &Peer{}, ErrPeerNotFound
	}
	p := data.(*Peer)
	return p, nil
}

func (s *memDBStore) On(event string, handler func(*Peer)) func() {
	return s.events.Subscribe(event, func(ev events.Event) {
		handler(ev.Entry.(*Peer))
	})
}
