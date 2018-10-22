package topics

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/weaveworks/mesh"
)

type memDBStore struct {
	db         *memdb.MemDB
	topicIndex *topicIndexer
	gossip     mesh.Gossip
}

var (
	ErrRetainedMessageNotFound = errors.New("retained message not found")
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

type Router interface {
	NewGossip(channel string, gossiper mesh.Gossiper) (mesh.Gossip, error)
}

type ByteSliceIndexer struct {
	i memdb.StringFieldIndex
}

func (b *ByteSliceIndexer) FromArgs(opts ...interface{}) ([]byte, error) {
	return b.i.FromArgs(opts...)
}

func (b *ByteSliceIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	message := obj.(*RetainedMessage)
	return true, append(message.GetTopic(), '\x00'), nil
}

func NewMemDBStore(router Router) (*memDBStore, error) {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"messages": &memdb.TableSchema{
				Name: "messages",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:         "id",
						AllowMissing: false,
						Unique:       true,
						Indexer: &memdb.StringFieldIndex{
							Field: "Id",
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
	}
	gossip, err := router.NewGossip("mqtt-topics", s)
	if err != nil {
		return nil, err
	}
	s.gossip = gossip
	return s, nil
}

type topicIndexer struct {
	root *Node
}

func TenantTopicIndexer() *topicIndexer {
	return &topicIndexer{
		root: NewNode("_root", "_all"),
	}
}

func (t *topicIndexer) Lookup(tenant string, pattern []byte) (RetainedMessageList, error) {
	var vals RetainedMessageList
	topic := NewTopic(pattern)
	t.root.Apply(tenant, topic, func(node *Node) bool {
		if node.Message != nil {
			vals.RetainedMessages = append(vals.RetainedMessages, node.Message)
		}
		return false
	})
	return vals, nil
}

func (s *topicIndexer) Index(message *RetainedMessage) error {
	topic := NewTopic(message.GetTopic())
	node := s.root.Upsert(message.GetTenant(), topic)
	node.Message = message
	return nil
}
func (m *memDBStore) do(write bool, f func(*memdb.Txn) error) error {
	tx := m.db.Txn(write)
	defer tx.Abort()
	return f(tx)
}
func (m *memDBStore) read(f func(*memdb.Txn) error) error {
	return m.do(false, f)
}
func (m *memDBStore) write(f func(*memdb.Txn) error) error {
	return m.do(true, f)
}

func (m *memDBStore) first(tx *memdb.Txn, index string, value ...interface{}) (*RetainedMessage, error) {
	var ok bool
	var res *RetainedMessage
	data, err := tx.First("messages", index, value...)
	if err != nil {
		return res, err
	}
	res, ok = data.(*RetainedMessage)
	if !ok {
		return res, errors.New("invalid type fetched")
	}
	if res.IsRemoved() {
		return nil, ErrRetainedMessageNotFound
	}
	return res, nil
}
func (m *memDBStore) all(tx *memdb.Txn, index string, value ...interface{}) (RetainedMessageList, error) {
	var set RetainedMessageList
	iterator, err := tx.Get("messages", index, value...)
	if err != nil {
		return set, err
	}
	for {
		data := iterator.Next()
		if data == nil {
			return set, nil
		}
		res, ok := data.(*RetainedMessage)
		if !ok {
			return set, errors.New("invalid type fetched")
		}
		if res.IsAdded() {
			set.RetainedMessages = append(set.RetainedMessages, res)
		}
	}
}

func (m *memDBStore) ByID(id string) (*RetainedMessage, error) {
	var res *RetainedMessage
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		return
	})
}
func (m *memDBStore) All() (RetainedMessageList, error) {
	var res RetainedMessageList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "id")
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (RetainedMessageList, error) {
	var res RetainedMessageList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) ByTopicPattern(tenant string, pattern []byte) (RetainedMessageList, error) {
	return m.topicIndex.Lookup(tenant, pattern)
}
func (m *memDBStore) Create(message *RetainedMessage) error {
	if message.Id == "" {
		id, err := MakeTopicID(message.Tenant, message.Topic)
		if err != nil {
			return err
		}
		message.Id = id
	}
	message.LastUpdated = now()
	err := m.topicIndex.Index(message)
	if err != nil {
		return err
	}
	defer m.gossip.GossipBroadcast(&RetainedMessageList{
		RetainedMessages: []*RetainedMessage{message},
	})
	return m.insert(message)
}
func (m *memDBStore) insert(message *RetainedMessage) error {
	if message.IsAdded() {
		err := m.topicIndex.Index(message)
		if err != nil {
			return err
		}
	}
	if message.IsRemoved() {
		message.Payload = nil
		err := m.topicIndex.Index(message)
		if err != nil {
			return err
		}
	}
	return m.write(func(tx *memdb.Txn) error {
		err := tx.Insert("messages", message)
		if err != nil {
			return err
		}
		tx.Commit()
		return nil
	})
}
