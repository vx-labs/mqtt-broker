package state

import (
	"errors"

	"github.com/golang/protobuf/proto"
)

// MockedBackend implements Backend interface using a map.
// This implementation is _not_ thread-safe.
type MockedBackend map[string]*MockedEntry

var _ Backend = &MockedBackend{}

var ErrMockedEntryNotFound = errors.New("mocked entry not found")

func (m MockedBackend) EntryByID(id string) (Entry, error) {
	e, ok := m[id]
	if !ok {
		return nil, ErrMockedEntryNotFound
	}
	return e, nil
}

func (m MockedBackend) InsertEntries(e EntrySet) error {
	e.Range(func(idx int, entry Entry) {
		mockedEntry := entry.(*MockedEntry)
		m[mockedEntry.GetID()] = mockedEntry
	})
	return nil
}
func (m MockedBackend) InsertEntry(entry Entry) error {
	mockedEntry := entry.(*MockedEntry)
	m[mockedEntry.GetID()] = mockedEntry
	return nil
}
func (m MockedBackend) DecodeSet(buf []byte) (EntrySet, error) {
	set := &MockedEntryList{}
	return set, proto.Unmarshal(buf, set)
}

func (m MockedBackend) DeleteEntry(e Entry) error {
	delete(m, e.GetID())
	return nil
}
func (m MockedBackend) Dump() EntrySet {
	set := &MockedEntryList{}
	for _, e := range m {
		set.MockedEntries = append(set.MockedEntries, e)
	}
	return set
}
