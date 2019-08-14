package peers

import (
	"errors"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/crdt"

	"github.com/vx-labs/mqtt-broker/events"

	memdb "github.com/hashicorp/go-memdb"
)

const (
	peerTable = "peers"
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
	pb.Metadata
}
type PeerStore interface {
	ByID(id string) (Peer, error)
	ByService(name string) (SubscriptionSet, error)
	EndpointsByService(name string) ([]*pb.NodeService, error)
	All() (SubscriptionSet, error)
	Exists(id string) bool
	Upsert(p Peer) error
	Delete(id string) error
	On(event string, handler func(Peer)) func()
}
type Channel interface {
	Broadcast(b []byte)
}
type memDBStore struct {
	db      *memdb.MemDB
	events  *events.Bus
	channel Channel
}

func NewPeerStore(mesh types.GossipServiceLayer) (*memDBStore, error) {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			peerTable: {
				Name: peerTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name: "id",
						Indexer: &memdb.StringFieldIndex{
							Field: "ID",
						},
						Unique:       true,
						AllowMissing: false,
					},
					"services": {
						Name: "services",
						Indexer: &memdb.StringSliceFieldIndex{
							Field: "Services",
						},
						Unique:       false,
						AllowMissing: true,
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
func (m *memDBStore) all(tx *memdb.Txn, index string, value ...interface{}) (SubscriptionSet, error) {
	var set SubscriptionSet
	iterator, err := tx.Get(peerTable, index, value...)
	if err != nil {
		return set, err
	}
	for {
		data := iterator.Next()
		if data == nil {
			return set, nil
		}
		res, ok := data.(Peer)
		if !ok {
			return set, errors.New("invalid type fetched")
		}
		if crdt.IsEntryAdded(&res) {
			set = append(set, res)
		}
	}
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
func (m *memDBStore) All() (SubscriptionSet, error) {
	var set SubscriptionSet
	var err error
	return set, m.read(func(tx *memdb.Txn) error {
		set, err = m.all(tx, "id")
		if err != nil {
			return err
		}
		return nil
	})
}
func (m *memDBStore) ByService(service string) (SubscriptionSet, error) {
	var res SubscriptionSet
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "services", service)
		return
	})
}

func (m *memDBStore) EndpointsByService(name string) ([]*pb.NodeService, error) {
	peers, err := m.ByService(name)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.NodeService, 0)
	for _, peer := range peers {
		for _, service := range peer.HostedServices {
			if service.ID == name {
				out = append(out, service)
			}
		}
	}
	return out, nil
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
	defer m.emitPeerEvent(message)
	err := m.write(func(tx *memdb.Txn) error {
		err := tx.Insert(peerTable, message)
		if err != nil {
			return err
		}
		tx.Commit()
		return nil
	})
	if err == nil {
		buf, err := proto.Marshal(&pb.PeerMetadataList{
			Metadatas: []*pb.Metadata{
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
	data, err := tx.First(peerTable, idx, id)
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
