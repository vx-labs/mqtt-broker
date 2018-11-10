package subscriptions

import (
	"errors"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-broker/state"
)

const table = "subscriptions"

type Store interface {
	ByTopic(tenant string, pattern []byte) (*SubscriptionList, error)
	ByID(id string) (*Subscription, error)
	All() (SubscriptionList, error)
	ByPeer(peer uint64) (SubscriptionList, error)
	BySession(id string) (SubscriptionList, error)
	Sessions() ([]string, error)
	Create(message *Subscription) error
	Delete(id string) error
	On(event string, handler func(*Subscription)) func()
}

const (
	SubscriptionCreated string = "subscription_created"
	SubscriptionDeleted string = "subscription_deleted"
)

type memDBStore struct {
	db           *memdb.MemDB
	state        *state.Store
	patternIndex *topicIndexer
	events       *events.Bus
}

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)
var now = func() int64 {
	return time.Now().UnixNano()
}

func NewMemDBStore(router state.Router) (Store, error) {
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
						Indexer:      &memdb.UintFieldIndex{Field: "Peer"},
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
		events:       events.NewEventBus(),
	}
	state, err := state.NewStore("mqtt-subscriptions", s, router)
	if err != nil {
		return nil, err
	}
	s.state = state
	return s, nil
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
func (t *topicIndexer) Lookup(tenant string, pattern []byte) (*SubscriptionList, error) {
	set := t.root.Select(tenant, nil, Topic(pattern)).Filter(func(s *Subscription) bool {
		return s.IsAdded()
	})
	return &set, nil
}

func (s *topicIndexer) Index(subscription *Subscription) error {
	s.root.Insert(
		Topic(subscription.Pattern),
		subscription.Tenant,
		subscription,
	)
	return nil
}

func (m *memDBStore) All() (SubscriptionList, error) {
	var set SubscriptionList
	var err error
	return set, m.read(func(tx *memdb.Txn) error {
		set, err = m.all(tx, "id")
		if err != nil {
			return err
		}
		return nil
	})
}

func (m *memDBStore) Gossiper() *state.Store {
	return m.state
}
func (m *memDBStore) ByID(id string) (*Subscription, error) {
	var res *Subscription
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		if res.IsRemoved() {
			return ErrSubscriptionNotFound
		}
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (SubscriptionList, error) {
	var res SubscriptionList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) BySession(session string) (SubscriptionList, error) {
	var res SubscriptionList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "session", session)
		return
	})
}
func (m *memDBStore) ByPeer(peer uint64) (SubscriptionList, error) {
	var res SubscriptionList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "peer", peer)
		return
	})
}
func (m *memDBStore) ByTopic(tenant string, pattern []byte) (*SubscriptionList, error) {
	return m.patternIndex.Lookup(tenant, pattern)
}
func (m *memDBStore) Sessions() ([]string, error) {
	var res SubscriptionList
	err := m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "session")
		return
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, len(res.Subscriptions))
	for idx := range res.Subscriptions {
		out[idx] = res.Subscriptions[idx].SessionID
	}
	return out, nil
}

func (m *memDBStore) insert(messages []*Subscription) error {
	return m.write(func(tx *memdb.Txn) error {
		for _, message := range messages {
			if message.IsAdded() {
				m.events.Emit(events.Event{
					Key:   SubscriptionCreated,
					Entry: message,
				})
				m.events.Emit(events.Event{
					Key:   SubscriptionCreated + "/" + message.SessionID,
					Entry: message,
				})
				err := m.patternIndex.Index(message)
				if err != nil {
					return err
				}
			}
			if message.IsRemoved() {
				m.events.Emit(events.Event{
					Key:   SubscriptionDeleted,
					Entry: message,
				})
				m.events.Emit(events.Event{
					Key:   SubscriptionDeleted + "/" + message.SessionID,
					Entry: message,
				})
				m.patternIndex.Remove(message.Tenant, message.ID, message.Pattern)
			}
			err := tx.Insert(table, message)
			if err != nil {
				return err
			}
		}
		tx.Commit()
		return nil
	})
}
func (m *memDBStore) Delete(id string) error {
	sub, err := m.ByID(id)
	if err != nil {
		return err
	}
	sub.LastDeleted = now()
	return m.state.Upsert(sub)
}
func (m *memDBStore) Create(message *Subscription) error {
	message.LastAdded = now()
	return m.state.Upsert(message)
}

func (s *memDBStore) On(event string, handler func(*Subscription)) func() {
	return s.events.Subscribe(event, func(ev events.Event) {
		handler(ev.Entry.(*Subscription))
	})
}
