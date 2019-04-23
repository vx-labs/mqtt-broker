package topics

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/vx-labs/mqtt-broker/broker/cluster"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-broker/state"
)

type memDBStore struct {
	db         *memdb.MemDB
	state      *state.Store
	topicIndex *topicIndexer
	events     *events.Bus
}
type Channel interface {
	Broadcast([]byte)
}

var (
	ErrRetainedMessageNotFound = errors.New("retained message not found")
)

const (
	RetainedMessageCreated string = "retained_message_created"
	RetainedMessageDeleted        = "retained_message_deleted"
	table                         = "messages"
)

func MakeTopicID(tenant string, topic []byte) (string, error) {
	hash := sha1.New()
	_, err := hash.Write([]byte(tenant))
	if err != nil {
		return "", err
	}
	_, err = hash.Write(topic)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

var now = func() int64 {
	return time.Now().UnixNano()
}

func NewMemDBStore(mesh cluster.Mesh) (*memDBStore, error) {
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
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	s := &memDBStore{
		db:         db,
		topicIndex: TenantTopicIndexer(),
		events:     events.NewEventBus(),
	}
	state, err := state.NewStore("mqtt-topics", mesh, s)
	if err != nil {
		return nil, err
	}
	s.state = state
	return s, nil
}

func (m *memDBStore) ByID(id string) (*Metadata, error) {
	var res *Metadata
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		if res.IsRemoved() {
			return ErrRetainedMessageNotFound
		}
		return
	})
}
func (m *memDBStore) All() (RetainedMessageMetadataList, error) {
	var res RetainedMessageMetadataList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "id")
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (RetainedMessageMetadataList, error) {
	var res RetainedMessageMetadataList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) ByTopicPattern(tenant string, pattern []byte) (RetainedMessageMetadataList, error) {
	set, err := m.topicIndex.Lookup(tenant, pattern)
	if err != nil {
		return RetainedMessageMetadataList{}, err
	}
	return set.Filter(func(m *Metadata) bool {
		return len(m.Payload) > 0
	}), nil
}
func (m *memDBStore) Create(message *Metadata) error {
	if message.ID == "" {
		id, err := MakeTopicID(message.Tenant, message.Topic)
		if err != nil {
			return err
		}
		message.ID = id
	}
	message.LastAdded = now()
	err := m.topicIndex.Index(message)
	if err != nil {
		return err
	}
	return m.state.Upsert(message)
}
func (m *memDBStore) insert(messages []*Metadata) error {
	return m.write(func(tx *memdb.Txn) error {
		for _, message := range messages {
			if message.IsAdded() {
				m.events.Emit(events.Event{
					Key:   RetainedMessageCreated,
					Entry: message,
				})
				err := m.topicIndex.Index(message)
				if err != nil {
					return err
				}
			}
			if message.IsRemoved() {
				m.events.Emit(events.Event{
					Key:   RetainedMessageDeleted,
					Entry: message,
				})
				message.Payload = nil
				err := m.topicIndex.Index(message)
				if err != nil {
					return err
				}
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
func (s *memDBStore) On(event string, handler func(*Metadata)) func() {
	return s.events.Subscribe(event, func(ev events.Event) {
		handler(ev.Entry.(*Metadata))
	})
}
