package sessions

import (
	"errors"
	"log"
	"time"

	"github.com/vx-labs/mqtt-broker/crdt"

	"github.com/vx-labs/mqtt-broker/broker/cluster"
	"github.com/vx-labs/mqtt-broker/events"

	memdb "github.com/hashicorp/go-memdb"
)

const (
	defaultTenant = "_default"
)
const (
	SessionCreated string = "session_created"
	SessionDeleted string = "session_deleted"
)

type Channel interface {
	Broadcast([]byte)
}

var (
	ErrSessionNotFound = errors.New("session not found")
)

var now = func() int64 {
	return time.Now().UnixNano()
}

type SessionWrapper struct {
	Session
	Close func() error
}
type SessionStore interface {
	ByID(id string) (SessionWrapper, error)
	ByClientID(id string) (SessionSet, error)
	ByPeer(peer string) (SessionSet, error)
	All() (SessionSet, error)
	Exists(id string) bool
	Delete(id, reason string) error
	Upsert(sess SessionWrapper, closer func() error) error
	On(event string, handler func(SessionWrapper)) func()
}

type memDBStore struct {
	db      *memdb.MemDB
	events  *events.Bus
	channel Channel
}

func NewSessionStore(mesh cluster.Mesh) (SessionStore, error) {
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
					"client_id": {
						Name: "client_id",
						Indexer: &memdb.StringFieldIndex{
							Field: "ClientID",
						},
						Unique:       false,
						AllowMissing: false,
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
		panic(err)
	}
	s := &memDBStore{
		db:     db,
		events: events.NewEventBus(),
	}
	s.channel, err = mesh.AddState("mqtt-sessions", s)
	go func() {
		for range time.Tick(1 * time.Hour) {
			err := s.runGC()
			if err != nil {
				log.Printf("WARN: failed to GC sessions: %v", err)
			}
		}
	}()
	return s, err
}
func (s *memDBStore) all() SessionSet {
	sessionList := make(SessionSet, 0)
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
			sess := payload.(SessionWrapper)
			sessionList = append(sessionList, sess)
		}
	})
	return sessionList
}

func (s *memDBStore) Exists(id string) bool {
	_, err := s.ByID(id)
	return err == nil
}
func (s *memDBStore) ByID(id string) (SessionWrapper, error) {
	var session SessionWrapper
	return session, s.read(func(tx *memdb.Txn) error {
		sess, err := s.first(tx, "id", id)
		if err != nil {
			return err
		}
		if crdt.IsEntryRemoved(&sess) {
			return ErrSessionNotFound
		}
		session = sess
		return nil
	})
}
func (s *memDBStore) ByClientID(id string) (SessionSet, error) {
	var sessionList SessionSet
	return sessionList, s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("sessions", "client_id", id)
		if err != nil || iterator == nil {
			return ErrSessionNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(SessionWrapper)
			if crdt.IsEntryAdded(&sess) {
				sessionList = append(sessionList, sess)
			}
		}
	})
}
func (s *memDBStore) All() (SessionSet, error) {
	return s.all().Filter(func(s SessionWrapper) bool {
		return crdt.IsEntryAdded(&s)
	}), nil
}

func (s *memDBStore) ByPeer(peer string) (SessionSet, error) {
	sessionList := make(SessionSet, 0)
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
			sess := payload.(SessionWrapper)
			if crdt.IsEntryAdded(&sess) {
				sessionList = append(sessionList, sess)
			}
		}
	})
}

func (s *memDBStore) Upsert(sess SessionWrapper, closer func() error) error {
	sess.LastAdded = now()
	sess.Close = closer
	if sess.Tenant == "" {
		sess.Tenant = defaultTenant
	}
	return s.insert(sess)
}
func (s *memDBStore) emitSessionEvent(sess SessionWrapper) {
	if crdt.IsEntryAdded(&sess) {
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   SessionCreated,
		})
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   SessionCreated + "/" + sess.ID,
		})
	}
	if crdt.IsEntryRemoved(&sess) {
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   SessionDeleted,
		})
		s.events.Emit(events.Event{
			Entry: sess,
			Key:   SessionDeleted + "/" + sess.ID,
		})
	}
}
func (s *memDBStore) insert(sess SessionWrapper) error {
	return s.write(func(tx *memdb.Txn) error {
		defer s.emitSessionEvent(sess)
		return tx.Insert("sessions", sess)
	})
}
func (s *memDBStore) Delete(id, reason string) error {
	sess, err := s.ByID(id)
	if err != nil {
		return err
	}
	sess.ClosureReason = reason
	sess.LastDeleted = now()
	return s.insert(sess)
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

func (s *memDBStore) first(tx *memdb.Txn, idx, id string) (SessionWrapper, error) {
	data, err := tx.First("sessions", idx, id)
	if err != nil || data == nil {
		return SessionWrapper{}, ErrSessionNotFound
	}
	sess := data.(SessionWrapper)
	return sess, nil
}

func (s *memDBStore) On(event string, handler func(SessionWrapper)) func() {
	return s.events.Subscribe(event, func(ev events.Event) {
		handler(ev.Entry.(SessionWrapper))
	})
}
