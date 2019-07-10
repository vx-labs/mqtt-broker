package subscriptions

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/crdt"
	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/cluster"
)

const table = "subscriptions"

type Subscription struct {
	Metadata
	Sender func(context.Context, packet.Publish) error
}

type Channel interface {
	Broadcast([]byte)
}

type RemoteSender func(host string, session string, publish packet.Publish) error

type Store interface {
	ByTopic(tenant string, pattern []byte) (SubscriptionSet, error)
	ByID(id string) (Subscription, error)
	All() (SubscriptionSet, error)
	ByPeer(peer string) (SubscriptionSet, error)
	BySession(id string) (SubscriptionSet, error)
	Sessions() ([]string, error)
	Create(message Subscription, sender func(context.Context, packet.Publish) error) error
	Delete(id string) error
}

const (
	SubscriptionCreated string = "subscription_created"
	SubscriptionDeleted string = "subscription_deleted"
)

type memDBStore struct {
	db           *memdb.MemDB
	patternIndex *topicIndexer
	channel      Channel
	sender       RemoteSender
}

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)
var now = func() int64 {
	return time.Now().UnixNano()
}

func NewMemDBStore(mesh cluster.ServiceLayer, sender RemoteSender) (Store, error) {
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
		return nil, err
	}

	s := &memDBStore{
		db:           db,
		patternIndex: TenantTopicIndexer(),
	}
	s.channel, err = mesh.AddState("mqtt-subscriptions", s)
	go func() {
		for range time.Tick(1 * time.Hour) {
			err := s.runGC()
			if err != nil {
				log.Printf("WARN: failed to GC sessions: %v", err)
			}
		}
	}()
	s.sender = sender
	return s, err
}

type topicIndexer struct {
	root *INode
}

func TenantTopicIndexer() *topicIndexer {
	return &topicIndexer{
		root: NewINode(),
	}
}
func (t *topicIndexer) Remove(tenant, id string, pattern []byte) error {
	return t.root.Remove(tenant, id, Topic(pattern))
}
func (t *topicIndexer) Lookup(tenant string, pattern []byte) (SubscriptionSet, error) {
	set := t.root.Select(tenant, nil, Topic(pattern)).Filter(func(s Subscription) bool {
		return crdt.IsEntryAdded(&s)
	})
	return set, nil
}

func (s *topicIndexer) Index(subscription Subscription) error {
	s.root.Insert(
		Topic(subscription.Pattern),
		subscription.Tenant,
		subscription,
	)
	return nil
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

func (m *memDBStore) ByID(id string) (Subscription, error) {
	var res Subscription
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		if crdt.IsEntryRemoved(&res) {
			return ErrSubscriptionNotFound
		}
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (SubscriptionSet, error) {
	var res SubscriptionSet
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) BySession(session string) (SubscriptionSet, error) {
	var res SubscriptionSet
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "session", session)
		return
	})
}
func (m *memDBStore) ByPeer(peer string) (SubscriptionSet, error) {
	var res SubscriptionSet
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "peer", peer)
		return
	})
}
func (m *memDBStore) ByTopic(tenant string, pattern []byte) (SubscriptionSet, error) {
	return m.patternIndex.Lookup(tenant, pattern)
}
func (m *memDBStore) Sessions() ([]string, error) {
	var res SubscriptionSet
	err := m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "session")
		return
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, len(res))
	for idx := range res {
		out[idx] = res[idx].SessionID
	}
	return out, nil
}

func (s *memDBStore) emitSubscriptionEvent(sess Subscription) {
	if crdt.IsEntryAdded(&sess) {
		s.patternIndex.Index(sess)
	}
	if crdt.IsEntryRemoved(&sess) {
		s.patternIndex.Remove(sess.Tenant, sess.ID, sess.Pattern)
	}
}

func (m *memDBStore) insert(message Subscription) error {
	defer m.emitSubscriptionEvent(message)
	err := m.write(func(tx *memdb.Txn) error {
		err := tx.Insert(table, message)
		if err != nil {
			return err
		}
		tx.Commit()
		return nil
	})
	if err == nil {
		buf, err := proto.Marshal(&SubscriptionMetadataList{
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
func (s *memDBStore) Create(sess Subscription, closer func(context.Context, packet.Publish) error) error {
	sess.LastAdded = now()
	sess.Sender = closer
	return s.insert(sess)
}
