package sessions

import (
	"errors"
	"time"

	"github.com/hashicorp/go-memdb"
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
	ById(id string) (*Session, error)
	ByPeer(peer uint64) (SessionList, error)
	All() (SessionList, error)
	Exists(id string) bool
	Upsert(sess *Session) error
	Delete(id string) error
	ComputeDelta(set *SessionList) (delta *SessionList)
	ApplyDelta(set *SessionList)
	DumpState() *SessionList
}

type MemDBStore struct {
	state  *memdb.MemDB
	gossip mesh.Gossip
}

type Router interface {
	NewGossip(channel string, gossiper mesh.Gossiper) (mesh.Gossip, error)
}

func NewSessionStore(router Router) SessionStore {
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
	s := &MemDBStore{
		state: db,
	}
	gossip, err := router.NewGossip("mqtt-sessions", s)
	if err != nil {
		panic(err)
	}
	s.gossip = gossip
	return s
}
func (s *MemDBStore) DumpState() *SessionList {
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
func (s *MemDBStore) ComputeDelta(set *SessionList) *SessionList {
	delta := SessionList{}
	set.Apply(func(remote *Session) {
		local, err := s.ById(remote.ID)
		if err != nil {
			// Session not found in our store, add it to delta
			delta.Sessions = append(delta.Sessions, remote)
		} else {
			if IsOutdated(local, remote) {
				delta.Sessions = append(delta.Sessions, remote)
			}
		}
	})
	return &delta
}
func (s *MemDBStore) ApplyDelta(set *SessionList) {
	set.Apply(func(remote *Session) {
		s.insert(remote)
	})
}
func (s *MemDBStore) Exists(id string) bool {
	_, err := s.ById(id)
	return err == nil
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
	return s.DumpState().Filter(func(s *Session) bool {
		return s.IsAdded()
	}), nil
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
			sess := payload.(*Session)
			if sess.IsAdded() {
				sessionList.Sessions = append(sessionList.Sessions, sess)
			}
		}
	})
}

func (s *MemDBStore) Upsert(sess *Session) error {
	sess.LastUpdated = now()
	if sess.Tenant == "" {
		sess.Tenant = defaultTenant
	}
	defer s.gossip.GossipBroadcast(&SessionList{
		Sessions: []*Session{sess},
	})
	return s.insert(sess)
}
func (s *MemDBStore) insert(sess *Session) error {
	return s.write(func(tx *memdb.Txn) error {
		return tx.Insert("sessions", sess)
	})
}
func (s *MemDBStore) Delete(id string) error {
	sess, err := s.ById(id)
	if err != nil {
		return nil
	}
	sess.LastDeleted = now()
	defer s.gossip.GossipBroadcast(&SessionList{
		Sessions: []*Session{sess},
	})
	return s.insert(sess)
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

func (s *MemDBStore) first(tx *memdb.Txn, idx, id string) (*Session, error) {
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
