package peers

import (
	"errors"

	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"

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

type PeerStore interface {
	ByID(id string) (*pb.Peer, error)
	ByService(name string) (*pb.PeerList, error)
	EndpointsByService(name string) ([]*pb.NodeService, error)
	All() (*pb.PeerList, error)
	Exists(id string) bool
	Upsert(p *pb.Peer) error
	Update(id string, mutation func(peer pb.Peer) pb.Peer) error
	Delete(id string) error
	On(event string, handler func(*pb.Peer)) func()
}
type memDBStore struct {
	db     *memdb.MemDB
	events *events.Bus
}

func NewPeerStore() (*memDBStore, error) {
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
	return s, nil
}
func (m *memDBStore) all(tx *memdb.Txn, index string, value ...interface{}) (*pb.PeerList, error) {
	set := &pb.PeerList{Peers: []*pb.Peer{}}
	iterator, err := tx.Get(peerTable, index, value...)
	if err != nil {
		return set, err
	}
	for {
		data := iterator.Next()
		if data == nil {
			return set, nil
		}
		res, ok := data.(*pb.Peer)
		if !ok {
			return set, errors.New("invalid type fetched")
		}
		set.Peers = append(set.Peers, res)

	}
}

func (s *memDBStore) Exists(id string) bool {
	_, err := s.ByID(id)
	return err == nil
}
func (s *memDBStore) ByID(id string) (*pb.Peer, error) {
	var peer *pb.Peer
	return peer, s.read(func(tx *memdb.Txn) error {
		p, err := s.first(tx, "id", id)
		if err != nil {
			return err
		}
		peer = p
		return nil
	})
}
func (m *memDBStore) All() (*pb.PeerList, error) {
	var set *pb.PeerList
	var err error
	return set, m.read(func(tx *memdb.Txn) error {
		set, err = m.all(tx, "id")
		if err != nil {
			return err
		}
		return nil
	})
}
func (m *memDBStore) ByService(service string) (*pb.PeerList, error) {
	var res *pb.PeerList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "services", service)
		return
	})
}

func (m *memDBStore) EndpointsByService(name string) ([]*pb.NodeService, error) {
	set, err := m.ByService(name)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.NodeService, 0)
	for _, peer := range set.Peers {
		for _, service := range peer.HostedServices {
			if service.ID == name {
				out = append(out, service)
			}
		}
	}
	return out, nil
}

func (s *memDBStore) Upsert(peer *pb.Peer) error {
	defer s.emitPeerEvent(PeerCreated, peer)
	return s.insert(peer)
}
func (s *memDBStore) Update(id string, mutation func(peer pb.Peer) pb.Peer) error {
	tx := s.db.Txn(true)
	defer tx.Abort()
	peer, err := s.first(tx, "id", id)
	if err != nil {
		return err
	}
	mutated := mutation(*peer)
	err = tx.Insert(peerTable, &mutated)
	if err != nil {
		return err
	}
	defer s.emitPeerEvent(PeerUpdated, peer)
	tx.Commit()
	return nil
}

func (s *memDBStore) emitPeerEvent(event string, sess *pb.Peer) {
	s.events.Emit(events.Event{
		Entry: sess,
		Key:   event,
	})
	s.events.Emit(events.Event{
		Entry: sess,
		Key:   event + "/" + sess.ID,
	})
}

func (m *memDBStore) insert(message *pb.Peer) error {
	return m.write(func(tx *memdb.Txn) error {
		return tx.Insert(peerTable, message)
	})
}
func (s *memDBStore) Delete(id string) error {
	tx := s.db.Txn(true)
	defer tx.Abort()
	peer, err := s.first(tx, "id", id)
	if err != nil {
		return err
	}
	err = tx.Delete(peerTable, peer)
	if err != nil {
		return err
	}
	defer s.emitPeerEvent(PeerDeleted, peer)
	tx.Commit()
	return nil
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

func (s *memDBStore) first(tx *memdb.Txn, idx, id string) (*pb.Peer, error) {
	data, err := tx.First(peerTable, idx, id)
	if err != nil || data == nil {
		return nil, ErrPeerNotFound
	}
	p := data.(*pb.Peer)
	return p, nil
}

func (s *memDBStore) On(event string, handler func(*pb.Peer)) func() {
	return s.events.Subscribe(event, func(ev events.Event) {
		handler(ev.Entry.(*pb.Peer))
	})
}
