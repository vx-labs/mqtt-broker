package topics

import (
	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
)

var _ state.Backend = &memDBStore{}
var _ state.EntrySet = &RetainedMessageMetadataList{}

func (e *RetainedMessageMetadataList) Append(entry state.Entry) {
	sub := entry.(*Metadata)
	e.Metadatas = append(e.Metadatas, sub)
}
func (e *RetainedMessageMetadataList) Length() int {
	return len(e.Metadatas)
}
func (e *RetainedMessageMetadataList) New() state.EntrySet {
	return &RetainedMessageMetadataList{}
}
func (e *RetainedMessageMetadataList) AtIndex(idx int) state.Entry {
	return e.Metadatas[idx]
}
func (e *RetainedMessageMetadataList) Set(idx int, entry state.Entry) {
	sub := entry.(*Metadata)
	e.Metadatas[idx] = sub
}
func (e *RetainedMessageMetadataList) Range(f func(idx int, entry state.Entry)) {
	for idx, entry := range e.Metadatas {
		f(idx, entry)
	}
}

func (m *memDBStore) EntryByID(id string) (state.Entry, error) {
	var session *Metadata
	err := m.read(func(tx *memdb.Txn) error {
		sess, err := m.first(tx, "id", id)
		if err != nil {
			return err
		}
		session = sess
		return nil
	})
	return session, err
}

func (m *memDBStore) InsertEntries(entries state.EntrySet) error {
	set := entries.(*RetainedMessageMetadataList)
	return m.insert(set.Metadatas)
}
func (m *memDBStore) InsertEntry(entry state.Entry) error {
	sub := entry.(*Metadata)
	return m.insert([]*Metadata{
		sub,
	})
}

func (m *memDBStore) DecodeSet(buf []byte) (state.EntrySet, error) {
	set := &RetainedMessageMetadataList{}
	return set, proto.Unmarshal(buf, set)
}
func (m memDBStore) Set() state.EntrySet {
	return &RetainedMessageMetadataList{}
}
func (m memDBStore) Dump() state.EntrySet {
	return m.DumpRetainedMessages()
}
func (m memDBStore) DumpRetainedMessages() *RetainedMessageMetadataList {
	RetainedMessageList := RetainedMessageMetadataList{}
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
			sess := payload.(*Metadata)
			RetainedMessageList.Metadatas = append(RetainedMessageList.Metadatas, sess)
		}
	})
	return &RetainedMessageList
}

func (m *memDBStore) DeleteEntry(entry state.Entry) error {
	RetainedMessage := entry.(*Metadata)
	return m.write(func(tx *memdb.Txn) error {
		err := tx.Delete("messages", RetainedMessage)
		if err == nil {
			tx.Commit()
		}
		return err
	})
}
