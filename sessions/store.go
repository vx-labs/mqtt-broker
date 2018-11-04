package sessions

import (
	"errors"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
	"github.com/weaveworks/mesh"
)

const (
	defaultTenant = "_default"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

var now = func() int64 {
	return time.Now().UnixNano()
}

type SessionStore interface {
	ByID(id string) (*Session, error)
	ByPeer(peer uint64) (SessionList, error)
	All() (SessionList, error)
	Exists(id string) bool
	Upsert(sess *Session) error
	Delete(id string) error
}

type memDBStore struct {
	db    *memdb.MemDB
	state *state.Store
}

type Router interface {
	NewGossip(channel string, gossiper mesh.Gossiper) (mesh.Gossip, error)
}

func NewSessionStore(router Router) (SessionStore, error) {
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
	s := &memDBStore{
		db: db,
	}
	state, err := state.NewStore("mqtt-sessions", s, router)
	if err != nil {
		return nil, err
	}
	s.state = state
	return s, nil
}
func (s *memDBStore) DumpState() *SessionList {
	sessionList := SessionList{}
	s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("sessions", "id")
		if err != nil || iterator == nil {
			return ErrSessionNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*Session)
			sessionList.Sessions = append(sessionList.Sessions, sess)
		}
	})
	return &sessionList
}

func (s *memDBStore) Exists(id string) bool {
	_, err := s.ByID(id)
	return err == nil
}
func (s *memDBStore) ByID(id string) (*Session, error) {
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
func (s *memDBStore) All() (SessionList, error) {
	return s.DumpState().Filter(func(s *Session) bool {
		return s.IsAdded()
	}), nil
}

func (s *memDBStore) ByPeer(peer uint64) (SessionList, error) {
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
			sess := payload.(*Session)
			if sess.IsAdded() {
				sessionList.Sessions = append(sessionList.Sessions, sess)
			}
		}
	})
}

func (s *memDBStore) Upsert(sess *Session) error {
	sess.LastAdded = now()
	if sess.Tenant == "" {
		sess.Tenant = defaultTenant
	}
	return s.state.Upsert(sess)
}
func (s *memDBStore) insert(sessions []*Session) error {
	return s.write(func(tx *memdb.Txn) error {
		for _, sess := range sessions {
			tx.Insert("sessions", sess)
		}
		return nil
	})
}
func (s *memDBStore) Delete(id string) error {
	sess, err := s.ByID(id)
	if err != nil {
		return nil
	}
	sess.LastDeleted = now()
	return s.state.Upsert(sess)
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

func (s *memDBStore) first(tx *memdb.Txn, idx, id string) (*Session, error) {
	data, err := tx.First("sessions", idx, id)
	if err != nil || data == nil {
		return &Session{}, ErrSessionNotFound
	}
	sess := data.(*Session)
	if sess.IsRemoved() {
		return nil, ErrSessionNotFound
	}
	return sess, nil
}
