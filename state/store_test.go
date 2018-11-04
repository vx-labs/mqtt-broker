package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	store := Store{
		backend: &MockedBackend{
			"1": &MockedEntry{ID: "1", Data: "1", LastAdded: 1},
			"2": &MockedEntry{ID: "2", Data: "2", LastAdded: 3},
		},
	}
	remote := &MockedEntryList{
		MockedEntries: []*MockedEntry{
			&MockedEntry{
				ID: "1", LastAdded: 4, Data: "9",
			},
		},
	}
	delta := store.ComputeDelta(remote)

	require.Equal(t, 1, delta.Length())
	entry := delta.AtIndex(0)
	require.NotNil(t, entry)
	mockedEntry := entry.(*MockedEntry)
	require.Equal(t, "9", mockedEntry.Data)
	require.Equal(t, "1", mockedEntry.ID)

	store.ApplyDelta(delta)
	entry, err := store.backend.EntryByID("1")
	require.NoError(t, err)
	mockedEntry = entry.(*MockedEntry)
	require.Equal(t, "9", mockedEntry.Data)
	require.Equal(t, "1", mockedEntry.ID)

}
