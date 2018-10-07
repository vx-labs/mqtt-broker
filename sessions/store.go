package sessions

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-memdb"
)

const (
	defaultTenant = "_default"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type SessionStore interface {
	ById(id string) (*Session, error)
	ByPeer(peer uint64) (SessionList, error)
	All() (SessionList, error)
	Exists(id string) bool
	Upsert(sess *Session) error
	Delete(id string) error
}

type MemDBStore struct {
	state *memdb.MemDB
}

func NewSessionStore() SessionStore {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"sessions": {
				Name: "sessions",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name: "id",
						Indexer: &memdb.StringFieldIndex{
							Field: "ID",
						},
						Unique:       true,
						AllowMissing: false,
					},
					"tenant": {
						Name: "tenant",
						Indexer: &memdb.StringFieldIndex{
							Field: "Tenant",
						},
						Unique:       false,
						AllowMissing: false,
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
		panic(err)
	}
	return &MemDBStore{
		state: db,
	}
}

func (s *MemDBStore) Exists(id string) bool {
	_, err := s.ById(id)
	return err == nil
}
func (s *MemDBStore) BySubscription(topic []byte) (SessionList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (s *MemDBStore) ById(id string) (*Session, error) {
	var session *Session
	return session, s.read(func(tx *memdb.Txn) error {
		sess, err := s.first(tx, "id", id)
		if err != nil {
			return err
		}
		session = sess
		return nil
	})
}
func (s *MemDBStore) All() (SessionList, error) {
	var sessionList SessionList
	return sessionList, s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("sessions", "id")
		if err != nil || iterator == nil {
			return ErrSessionNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sessionList = append(sessionList, payload.(*Session))
		}
	})
}

func (s *MemDBStore) ByPeer(peer uint64) (SessionList, error) {
	var sessionList SessionList
	return sessionList, s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("sessions", "peer", peer)
		if err != nil || iterator == nil {
			return ErrSessionNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sessionList = append(sessionList, payload.(*Session))
		}
	})
}

func (s *MemDBStore) Upsert(sess *Session) error {
	if sess.Tenant == "" {
		sess.Tenant = defaultTenant
	}
	return s.write(func(tx *memdb.Txn) error {
		return tx.Insert("sessions", sess)
	})
}
func (s *MemDBStore) Delete(id string) error {
	return s.write(func(tx *memdb.Txn) error {
		return tx.Delete("sessions", &Session{ID: id})
	})
}

func (s *MemDBStore) read(statement func(tx *memdb.Txn) error) error {
	tx := s.state.Txn(false)
	return s.run(tx, statement)
}
func (s *MemDBStore) write(statement func(tx *memdb.Txn) error) error {
	tx := s.state.Txn(true)
	return s.run(tx, statement)
}
func (s *MemDBStore) run(tx *memdb.Txn, statement func(tx *memdb.Txn) error) error {
	defer tx.Abort()
	err := statement(tx)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (s *MemDBStore) update(tx *memdb.Txn, idx, id string, updater func(session *Session) (*Session, error)) error {
	iterator, err := tx.Get("sessions", idx, id)
	if err != nil {
		return err
	}
	for {
		payload := iterator.Next()
		if payload == nil {
			return nil
		}
		sess, err := updater(payload.(*Session))
		if err != nil {
			return err
		}
		tx.Insert("sessions", sess)
	}
}
func (s *MemDBStore) first(tx *memdb.Txn, idx, id string) (*Session, error) {
	data, err := tx.First("sessions", idx, id)
	if err != nil || data == nil {
		return &Session{}, ErrSessionNotFound
	}
	return data.(*Session), nil
}
