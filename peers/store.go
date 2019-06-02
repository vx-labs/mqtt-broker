package peers

import (
	"errors"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/crdt"

	"github.com/vx-labs/mqtt-broker/broker/cluster"
	"github.com/vx-labs/mqtt-broker/events"

	memdb "github.com/hashicorp/go-memdb"
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

type Peer struct {
	Metadata
}
type PeerStore interface {
	ByID(id string) (Peer, error)
	All() (SubscriptionSet, error)
	Exists(id string) bool
	Upsert(p Peer) error
	Delete(id string) error
	On(event string, handler func(Peer)) func()
}

type memDBStore struct {
	db      *memdb.MemDB
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
	s.channel, err = mesh.AddState("mqtt-peers", s)
	go func() {
		for range time.Tick(1 * time.Hour) {
			err := s.runGC()
			if err != nil {
				log.Printf("WARN: failed to GC peers: %v", err)
			}
		}
	}()
	return s, nil
}
func (s *memDBStore) all() SubscriptionSet {
	peerList := SubscriptionSet{}
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
			p := payload.(Peer)
			peerList = append(peerList, p)
		}
	})
	return peerList
}

func (s *memDBStore) Exists(id string) bool {
	_, err := s.ByID(id)
	return err == nil
}
func (s *memDBStore) ByID(id string) (Peer, error) {
	var peer Peer
	return peer, s.read(func(tx *memdb.Txn) error {
		p, err := s.first(tx, "id", id)
		if err != nil {
			return err
		}
		if crdt.IsEntryRemoved(&p) {
			return ErrPeerNotFound
		}
		peer = p
		return nil
	})
}
func (s *memDBStore) All() (SubscriptionSet, error) {
	return s.all().Filter(func(s Peer) bool {
		return crdt.IsEntryAdded(&s)
	}), nil
}

func (s *memDBStore) Upsert(sess Peer) error {
	sess.LastAdded = now()
	return s.insert(sess)
}

func (s *memDBStore) emitPeerEvent(sess Peer) {
	if crdt.IsEntryAdded(&sess) {
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   PeerCreated,
		})
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   PeerCreated + "/" + sess.ID,
		})
	}
	if crdt.IsEntryRemoved(&sess) {
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   PeerDeleted,
		})
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   PeerDeleted + "/" + sess.ID,
		})
	}
}

func (m *memDBStore) insert(message Peer) error {
	err := m.write(func(tx *memdb.Txn) error {
		m.emitPeerEvent(message)
		err := tx.Insert(table, message)
		if err != nil {
			return err
		}
		tx.Commit()
		return nil
	})
	if err == nil {
		buf, err := proto.Marshal(&PeerMetadataList{
			Metadatas: []*Metadata{
				&message.Metadata,
			},
		})
		if err != nil {
			return err
		}
		m.channel.Broadcast(buf)
	}
	return err
}
func (s *memDBStore) Delete(id string) error {
	sess, err := s.ByID(id)
	if err != nil {
		return err
	}
	sess.LastDeleted = now()
	return s.insert(sess)
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

func (s *memDBStore) first(tx *memdb.Txn, idx, id string) (Peer, error) {
	data, err := tx.First(table, idx, id)
	if err != nil || data == nil {
		return Peer{}, ErrPeerNotFound
	}
	p := data.(Peer)
	return p, nil
}

func (s *memDBStore) On(event string, handler func(Peer)) func() {
	return s.events.Subscribe(event, func(ev events.Event) {
		handler(ev.Entry.(Peer))
	})
}
