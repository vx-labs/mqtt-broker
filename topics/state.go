package topics

import (
	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
)

var _ state.Backend = &memDBStore{}
var _ state.EntrySet = &RetainedMessageList{}

func (e *RetainedMessageList) Append(entry state.Entry) {
	sub := entry.(*RetainedMessage)
	e.RetainedMessages = append(e.RetainedMessages, sub)
}
func (e *RetainedMessageList) Length() int {
	return len(e.RetainedMessages)
}
func (e *RetainedMessageList) New() state.EntrySet {
	return &RetainedMessageList{}
}
func (e *RetainedMessageList) AtIndex(idx int) state.Entry {
	return e.RetainedMessages[idx]
}
func (e *RetainedMessageList) Set(idx int, entry state.Entry) {
	sub := entry.(*RetainedMessage)
	e.RetainedMessages[idx] = sub
}
func (e *RetainedMessageList) Range(f func(idx int, entry state.Entry)) {
	for idx, entry := range e.RetainedMessages {
		f(idx, entry)
	}
}

func (m *memDBStore) EntryByID(id string) (state.Entry, error) {
	return m.ByID(id)
}

func (m *memDBStore) InsertEntries(entries state.EntrySet) error {
	set := entries.(*RetainedMessageList)
	return m.insert(set.RetainedMessages)
}
func (m *memDBStore) InsertEntry(entry state.Entry) error {
	sub := entry.(*RetainedMessage)
	return m.insert([]*RetainedMessage{
		sub,
	})
}

func (m *memDBStore) DecodeSet(buf []byte) (state.EntrySet, error) {
	set := &RetainedMessageList{}
	return set, proto.Unmarshal(buf, set)
}
func (m memDBStore) Dump() state.EntrySet {
	RetainedMessageList := RetainedMessageList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("messages", "id")
		if err != nil || iterator == nil {
			return nil
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*RetainedMessage)
			RetainedMessageList.RetainedMessages = append(RetainedMessageList.RetainedMessages, sess)
		}
	})
	return &RetainedMessageList
}

func (m *memDBStore) DeleteEntry(entry state.Entry) error {
	RetainedMessage := entry.(*RetainedMessage)
	return m.write(func(tx *memdb.Txn) error {
		err := tx.Delete("messages", RetainedMessage)
		if err == nil {
			tx.Commit()
		}
		return err
	})
}
