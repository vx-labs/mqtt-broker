package subscriptions

import (
	"errors"
	"log"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"github.com/vx-labs/mqtt-broker/subscriptions/pb"
	"github.com/vx-labs/mqtt-broker/subscriptions/topic"
	"github.com/vx-labs/mqtt-broker/subscriptions/tree"
	"go.uber.org/zap"

	"github.com/hashicorp/go-memdb"
)

const table = "subscriptions"

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type Store interface {
	ByTopic(tenant string, pattern []byte) (*pb.SubscriptionMetadataList, error)
	ByID(id string) (*pb.Subscription, error)
	All() (*pb.SubscriptionMetadataList, error)
	BySession(id string) (*pb.SubscriptionMetadataList, error)
	Create(message *pb.Subscription) error
	Delete(id string) error
}
type Channel interface {
	Broadcast([]byte)
}

type memDBStore struct {
	db           *memdb.MemDB
	patternIndex *topicIndexer
	channel      Channel
	logger       *zap.Logger
}

func NewSubscriptionStore(mesh types.GossipServiceLayer, logger *zap.Logger) Store {
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
		logger:       logger,
	}
	s.channel, err = mesh.AddState("mqtt-subscriptions", s)
	if err != nil {
		panic(err)
	}
	go func() {
		for range time.Tick(1 * time.Hour) {
			err := s.runGC()
			if err != nil {
				log.Printf("WARN: failed to GC sessions: %v", err)
			}
		}
	}()
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
		Subscriptions: set,
	}, nil
}

func (s *topicIndexer) Index(subscription *pb.Subscription) error {
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

func (m *memDBStore) ByID(id string) (*pb.Subscription, error) {
	var res *pb.Subscription
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
func (m *memDBStore) ByTopic(tenant string, pattern []byte) (*pb.SubscriptionMetadataList, error) {
	return m.patternIndex.Lookup(tenant, pattern)
}

func (m *memDBStore) insert(message *pb.Subscription) error {
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
	var subscription *pb.Subscription
	err := s.write(func(tx *memdb.Txn) error {
		var err error
		subscription, err = s.first(tx, "id", id)
		if err != nil {
			return err
		}
		subscription.LastDeleted = time.Now().UnixNano()
		return tx.Insert(table, subscription)
	})
	if err == nil {
		s.patternIndex.Index(subscription)
		buf, err := proto.Marshal(&pb.SubscriptionMetadataList{
			Subscriptions: []*pb.Subscription{
				subscription,
			},
		})
		if err != nil {
			return err
		}
		s.channel.Broadcast(buf)
	}
	return err
}
func (s *memDBStore) Create(sess *pb.Subscription) error {
	err := s.write(func(tx *memdb.Txn) error {
		sess.LastAdded = time.Now().UnixNano()
		return tx.Insert(table, sess)
	})
	if err == nil {
		s.patternIndex.Index(sess)
		buf, err := proto.Marshal(&pb.SubscriptionMetadataList{
			Subscriptions: []*pb.Subscription{
				sess,
			},
		})
		if err != nil {
			s.logger.Error("failed to marshal subscription in distributed state")
			return err
		}
		s.channel.Broadcast(buf)
	}
	return err

}
