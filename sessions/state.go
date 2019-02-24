package sessions

import (
	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
)

var _ state.Backend = &memDBStore{}
var _ state.EntrySet = &SessionList{}

func (e *SessionList) Append(entry state.Entry) {
	sub := entry.(*Session)
	e.Sessions = append(e.Sessions, sub)
}
func (e *SessionList) Length() int {
	return len(e.Sessions)
}
func (e *SessionList) New() state.EntrySet {
	return &SessionList{}
}
func (e *SessionList) AtIndex(idx int) state.Entry {
	return e.Sessions[idx]
}
func (e *SessionList) Set(idx int, entry state.Entry) {
	sub := entry.(*Session)
	e.Sessions[idx] = sub
}
func (e *SessionList) Range(f func(idx int, entry state.Entry)) {
	for idx, entry := range e.Sessions {
		f(idx, entry)
	}
}

func (m *memDBStore) EntryByID(id string) (state.Entry, error) {
	var session Session
	err := m.read(func(tx *memdb.Txn) error {
		sess, err := m.first(tx, "id", id)
		if err != nil {
			return err
		}
		session = sess
		return nil
	})
	return &session, err
}

func (m *memDBStore) InsertEntries(entries state.EntrySet) error {
	set := entries.(*SessionList)
	return m.insert(set.Sessions)
}
func (m *memDBStore) InsertEntry(entry state.Entry) error {
	sub := entry.(*Session)
	return m.insert([]*Session{
		sub,
	})
}

func (m *memDBStore) DecodeSet(buf []byte) (state.EntrySet, error) {
	set := &SessionList{}
	return set, proto.Unmarshal(buf, set)
}
func (m memDBStore) Set() state.EntrySet {
	return &SessionList{}
}
func (m memDBStore) Dump() state.EntrySet {
	sessionList := SessionList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("sessions", "id")
		if (err != nil && err != ErrSessionNotFound) || iterator == nil {
			return err
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(Session)
			sessionList.Sessions = append(sessionList.Sessions, &sess)
		}
	})
	return &sessionList
}

func (m *memDBStore) DeleteEntry(entry state.Entry) error {
	session := entry.(*Session)
	return m.write(func(tx *memdb.Txn) error {
		err := tx.Delete("sessions", *session)
		if err == nil {
			tx.Commit()
		}
		return err
	})
}
