package subscriptions

import (
	"errors"

	"github.com/vx-labs/mqtt-broker/subscriptions/pb"
	"github.com/vx-labs/mqtt-broker/subscriptions/topic"
	"github.com/vx-labs/mqtt-broker/subscriptions/tree"

	"github.com/hashicorp/go-memdb"
)

const table = "subscriptions"

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type Store interface {
	ByTopic(tenant string, pattern []byte) (*pb.SubscriptionMetadataList, error)
	ByID(id string) (*pb.Metadata, error)
	All() (*pb.SubscriptionMetadataList, error)
	ByPeer(peer string) (*pb.SubscriptionMetadataList, error)
	BySession(id string) (*pb.SubscriptionMetadataList, error)
	Create(message *pb.Metadata) error
	Delete(id string) error
}

type memDBStore struct {
	db           *memdb.MemDB
	patternIndex *topicIndexer
}

func NewMemDBStore() Store {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			table: &memdb.TableSchema{
				Name: table,
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:         "id",
						AllowMissing: false,
						Unique:       true,
						Indexer: &memdb.StringFieldIndex{
							Field: "ID",
						},
					},
					"tenant": &memdb.IndexSchema{
						Name:         "tenant",
						AllowMissing: false,
						Unique:       false,
						Indexer:      &memdb.StringFieldIndex{Field: "Tenant"},
					},
					"session": &memdb.IndexSchema{
						Name:         "session",
						AllowMissing: false,
						Unique:       false,
						Indexer:      &memdb.StringFieldIndex{Field: "SessionID"},
					},
					"peer": &memdb.IndexSchema{
						Name:         "peer",
						AllowMissing: false,
						Unique:       false,
						Indexer:      &memdb.StringFieldIndex{Field: "Peer"},
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	s := &memDBStore{
		db:           db,
		patternIndex: TenantTopicIndexer(),
	}
	return s
}

type topicIndexer struct {
	root *tree.INode
}

func TenantTopicIndexer() *topicIndexer {
	return &topicIndexer{
		root: tree.NewINode(),
	}
}
func (t *topicIndexer) Remove(tenant, id string, pattern []byte) error {
	return t.root.Remove(tenant, id, topic.Topic(pattern))
}
func (t *topicIndexer) Lookup(tenant string, pattern []byte) (*pb.SubscriptionMetadataList, error) {
	set := t.root.Select(tenant, nil, topic.Topic(pattern))
	return &pb.SubscriptionMetadataList{
		Metadatas: set,
	}, nil
}

func (s *topicIndexer) Index(subscription *pb.Metadata) error {
	s.root.Insert(
		topic.Topic(subscription.Pattern),
		subscription.Tenant,
		subscription,
	)
	return nil
}

func (m *memDBStore) All() (*pb.SubscriptionMetadataList, error) {
	var set *pb.SubscriptionMetadataList
	var err error
	return set, m.read(func(tx *memdb.Txn) error {
		set, err = m.all(tx, "id")
		if err != nil {
			return err
		}
		return nil
	})
}

func (m *memDBStore) ByID(id string) (*pb.Metadata, error) {
	var res *pb.Metadata
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (*pb.SubscriptionMetadataList, error) {
	var res *pb.SubscriptionMetadataList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) BySession(session string) (*pb.SubscriptionMetadataList, error) {
	var res *pb.SubscriptionMetadataList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "session", session)
		return
	})
}
func (m *memDBStore) ByPeer(peer string) (*pb.SubscriptionMetadataList, error) {
	var res *pb.SubscriptionMetadataList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "peer", peer)
		return
	})
}
func (m *memDBStore) ByTopic(tenant string, pattern []byte) (*pb.SubscriptionMetadataList, error) {
	return m.patternIndex.Lookup(tenant, pattern)
}

func (m *memDBStore) insert(message *pb.Metadata) error {
	err := m.write(func(tx *memdb.Txn) error {
		err := tx.Insert(table, message)
		if err != nil {
			return err
		}
		tx.Commit()
		return nil
	})
	if err == nil {
		return m.patternIndex.Index(message)
	}
	return err
}
func (s *memDBStore) Delete(id string) error {
	var oldSub *pb.Metadata
	err := s.write(func(tx *memdb.Txn) error {
		sub, err := tx.First(table, "id", id)
		if err != nil {
			return err
		}
		if sub == nil {
			return nil
		}
		oldSub = sub.(*pb.Metadata)
		return tx.Delete(table, oldSub)
	})
	if err == nil && oldSub != nil {
		return s.patternIndex.Remove(oldSub.Tenant, oldSub.ID, oldSub.Pattern)
	}
	return err
}
func (s *memDBStore) Create(sess *pb.Metadata) error {
	return s.insert(sess)
}
