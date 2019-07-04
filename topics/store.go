package topics

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/crdt"

	"github.com/vx-labs/mqtt-broker/cluster"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/events"
)

type memDBStore struct {
	db         *memdb.MemDB
	topicIndex *topicIndexer
	events     *events.Bus
	channel    Channel
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

type RetainedMessage struct {
	Metadata
}

func makeTopicID(tenant string, topic []byte) (string, error) {
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

func NewMemDBStore(mesh cluster.ServiceLayer) (*memDBStore, error) {
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
	s.channel, err = mesh.AddState("mqtt-topics", s)
	go func() {
		for range time.Tick(1 * time.Hour) {
			err := s.runGC()
			if err != nil {
				log.Printf("WARN: failed to GC sessions: %v", err)
			}
		}
	}()
	return s, nil
}

func (m *memDBStore) ByID(id string) (RetainedMessage, error) {
	var res RetainedMessage
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		if crdt.IsEntryRemoved(&res) {
			return ErrRetainedMessageNotFound
		}
		return
	})
}
func (m *memDBStore) All() (RetainedMessageSet, error) {
	var res RetainedMessageSet
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "id")
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (RetainedMessageSet, error) {
	var res RetainedMessageSet
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) ByTopicPattern(tenant string, pattern []byte) (RetainedMessageSet, error) {
	set, err := m.topicIndex.Lookup(tenant, pattern)
	if err != nil {
		return RetainedMessageSet{}, err
	}
	return set.Filter(func(m RetainedMessage) bool {
		return len(m.Payload) > 0
	}), nil
}
func (s *memDBStore) Create(sess RetainedMessage) error {
	sess.LastAdded = now()
	var err error
	if sess.ID == "" {
		sess.ID, err = makeTopicID(sess.Tenant, sess.Topic)
		if err != nil {
			return err
		}
	}
	err = s.topicIndex.Index(sess)
	if err != nil {
		return err
	}
	return s.insert(sess)
}
func (m *memDBStore) insert(message RetainedMessage) error {
	defer m.emitRetainedMessageEvent(message)
	err := m.write(func(tx *memdb.Txn) error {
		err := tx.Insert(table, message)
		if err != nil {
			return err
		}
		tx.Commit()
		return nil
	})
	if err == nil {
		buf, err := proto.Marshal(&RetainedMessageMetadataList{
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
func (s *memDBStore) On(event string, handler func(RetainedMessage)) func() {
	return s.events.Subscribe(event, func(ev events.Event) {
		handler(ev.Entry.(RetainedMessage))
	})
}

func (s *memDBStore) emitRetainedMessageEvent(sess RetainedMessage) {
	if crdt.IsEntryAdded(&sess) {
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   RetainedMessageCreated,
		})
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   RetainedMessageCreated + "/" + sess.ID,
		})
	}
	if crdt.IsEntryRemoved(&sess) {
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   RetainedMessageDeleted,
		})
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   RetainedMessageDeleted + "/" + sess.ID,
		})
	}
}
