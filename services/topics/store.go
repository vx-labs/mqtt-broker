package topics

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/crdt"
	"github.com/vx-labs/mqtt-broker/services/topics/pb"

	memdb "github.com/hashicorp/go-memdb"
)

type memDBStore struct {
	db         *memdb.MemDB
	topicIndex *topicIndexer
	events     chan []byte
}

var (
	ErrRetainedMessageNotFound = errors.New("retained message not found")
)

const (
	RetainedMessageCreated string = "retained_message_created"
	RetainedMessageDeleted        = "retained_message_deleted"
	table                         = "messages"
)

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

func NewMemDBStore() *memDBStore {
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
		panic(err)
	}
	s := &memDBStore{
		db:         db,
		topicIndex: TenantTopicIndexer(),
		events:     make(chan []byte),
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
func (s *memDBStore) Events() chan []byte {
	return s.events
}
func (s *memDBStore) notify(b []byte) {
	select {
	case s.events <- b:
	default:
	}
}

func (m *memDBStore) ByID(id string) (pb.RetainedMessage, error) {
	var res pb.RetainedMessage
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		if crdt.IsEntryRemoved(&res) {
			return ErrRetainedMessageNotFound
		}
		return
	})
}
func (m *memDBStore) All() (pb.RetainedMessageSet, error) {
	var res pb.RetainedMessageSet
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "id")
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (pb.RetainedMessageSet, error) {
	var res pb.RetainedMessageSet
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) ByTopicPattern(tenant string, pattern []byte) (pb.RetainedMessageSet, error) {
	set, err := m.topicIndex.Lookup(tenant, pattern)
	if err != nil {
		return pb.RetainedMessageSet{}, err
	}
	return set.Filter(func(m pb.RetainedMessage) bool {
		return len(m.Payload) > 0
	}), nil
}
func (s *memDBStore) Create(sess pb.RetainedMessage) error {
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
func (m *memDBStore) insert(message pb.RetainedMessage) error {
	err := m.write(func(tx *memdb.Txn) error {
		err := tx.Insert(table, message)
		if err != nil {
			return err
		}
		tx.Commit()
		return nil
	})
	if err == nil {
		buf, err := proto.Marshal(&pb.RetainedMessageMetadataList{
			RetainedMessages: []*pb.RetainedMessage{
				&message,
			},
		})
		if err != nil {
			return err
		}
		m.notify(buf)
	}
	return err
}
