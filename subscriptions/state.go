package subscriptions

import (
	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
)

var _ state.Backend = &memDBStore{}
var _ state.EntrySet = &SubscriptionList{}

func (e *SubscriptionList) Append(entry state.Entry) {
	sub := entry.(*Subscription)
	e.Subscriptions = append(e.Subscriptions, sub)
}
func (e *SubscriptionList) Length() int {
	return len(e.Subscriptions)
}
func (e *SubscriptionList) New() state.EntrySet {
	return &SubscriptionList{}
}
func (e *SubscriptionList) AtIndex(idx int) state.Entry {
	return e.Subscriptions[idx]
}
func (e *SubscriptionList) Set(idx int, entry state.Entry) {
	sub := entry.(*Subscription)
	e.Subscriptions[idx] = sub
}
func (e *SubscriptionList) Range(f func(idx int, entry state.Entry)) {
	for idx, entry := range e.Subscriptions {
		f(idx, entry)
	}
}

func (m *memDBStore) EntryByID(id string) (state.Entry, error) {
	return m.ByID(id)
}

func (m *memDBStore) InsertEntries(entries state.EntrySet) error {
	set := entries.(*SubscriptionList)
	return m.insert(set.Subscriptions)
}
func (m *memDBStore) InsertEntry(entry state.Entry) error {
	sub := entry.(*Subscription)
	return m.insert([]*Subscription{
		sub,
	})
}

func (m *memDBStore) DecodeSet(buf []byte) (state.EntrySet, error) {
	set := &SubscriptionList{}
	return set, proto.Unmarshal(buf, set)
}
func (m memDBStore) Dump() state.EntrySet {
	sessionList := SubscriptionList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("subscriptions", "id")
		if err != nil || iterator == nil {
			return nil
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*Subscription)
			sessionList.Subscriptions = append(sessionList.Subscriptions, sess)
		}
	})
	return &sessionList
}

func (m *memDBStore) DeleteEntry(entry state.Entry) error {
	session := entry.(*Subscription)
	session.LastDeleted = now()
	return m.write(func(tx *memdb.Txn) error {
		err := tx.Delete("subscriptions", session)
		if err == nil {
			tx.Commit()
		}
		return err
	})
}
