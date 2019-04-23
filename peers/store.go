package peers

import (
	"errors"
	"time"

	"github.com/vx-labs/mqtt-broker/broker/cluster"
	"github.com/vx-labs/mqtt-broker/events"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
)

const (
	table = "peers"
)
const (
	PeerCreated string = "peer_created"
	PeerUpdated string = "peer_updated"
	PeerDeleted string = "peer_deleted"
)

var (
	ErrPeerNotFound = errors.New("peer not found")
)

var now = func() int64 {
	return time.Now().UnixNano()
}

type PeerStore interface {
	ByID(id string) (*Metadata, error)
	ByMeshID(id string) (*Metadata, error)
	All() (PeerMetadataList, error)
	Exists(id string) bool
	Upsert(p *Metadata) error
	Delete(id string) error
	On(event string, handler func(*Metadata)) func()
	DumpPeers() *PeerMetadataList
}

type memDBStore struct {
	db      *memdb.MemDB
	state   *state.Store
	events  *events.Bus
	channel Channel
}
type Channel interface {
	Broadcast([]byte)
}

func NewPeerStore(mesh cluster.Mesh) (PeerStore, error) {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			table: {
				Name: table,
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
						Indexer:      &memdb.StringFieldIndex{Field: "MeshID"},
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
	return s, nil
}
func (s *memDBStore) all() *PeerMetadataList {
	peerList := PeerMetadataList{}
	s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(table, "id")
		if err != nil || iterator == nil {
			return ErrPeerNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			p := payload.(*Metadata)
			peerList.Peers = append(peerList.Peers, p)
		}
	})
	return &peerList
}

func (s *memDBStore) Exists(id string) bool {
	_, err := s.ByID(id)
	return err == nil
}
func (s *memDBStore) ByID(id string) (*Metadata, error) {
	var peer *Metadata
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
func (s *memDBStore) All() (PeerMetadataList, error) {
	return s.all().Filter(func(s *Metadata) bool {
		return s.IsAdded()
	}), nil
}

func (s *memDBStore) ByMeshID(id string) (*Metadata, error) {
	var peer *Metadata
	err := s.read(func(tx *memdb.Txn) error {
		data, err := tx.First(table, "meshID", id)
		if err != nil || data == nil {
			return ErrPeerNotFound
		}
		p := data.(*Metadata)
		peer = p
		return nil
	})
	return peer, err
}

func (s *memDBStore) Upsert(p *Metadata) error {
	p.LastAdded = now()
	return s.state.Upsert(p)
}
func (s *memDBStore) insert(peers []*Metadata) error {
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
			err := tx.Insert(table, p)
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

func (s *memDBStore) first(tx *memdb.Txn, idx, id string) (*Metadata, error) {
	data, err := tx.First(table, idx, id)
	if err != nil || data == nil {
		return &Metadata{}, ErrPeerNotFound
	}
	p := data.(*Metadata)
	return p, nil
}

func (s *memDBStore) On(event string, handler func(*Metadata)) func() {
	return s.events.Subscribe(event, func(ev events.Event) {
		handler(ev.Entry.(*Metadata))
	})
}
