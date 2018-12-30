package state

import (
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	t.Run("without delta", func(t *testing.T) {
		store := Store{
			backend: &MockedBackend{
				"1": &MockedEntry{ID: "1", Data: "1", LastAdded: 2},
				"2": &MockedEntry{ID: "2", Data: "2", LastAdded: 3},
			},
		}
		remote := &MockedEntryList{
			MockedEntries: []*MockedEntry{
				&MockedEntry{ID: "1", Data: "9", LastAdded: 1},
				&MockedEntry{ID: "2", Data: "1", LastAdded: 1},
			},
		}
		payload, err := proto.Marshal(remote)
		require.NoError(t, err)
		data, err := store.merge(payload)
		require.NoError(t, err)
		require.Nil(t, data)
	})
}
func TestComputeDelta(t *testing.T) {
	t.Run("without delta", func(t *testing.T) {
		store := Store{
			backend: &MockedBackend{
				"1": &MockedEntry{ID: "1", Data: "1", LastAdded: 2},
				"2": &MockedEntry{ID: "2", Data: "2", LastAdded: 3},
			},
		}
		remote := &MockedEntryList{
			MockedEntries: []*MockedEntry{
				&MockedEntry{ID: "1", Data: "9", LastAdded: 1},
				&MockedEntry{ID: "2", Data: "1", LastAdded: 1},
			},
		}
		delta := store.ComputeDelta(remote)
		require.Equal(t, 0, delta.Length())
	})
	t.Run("with delta", func(t *testing.T) {
		store := Store{
			backend: &MockedBackend{
				"1": &MockedEntry{ID: "1", Data: "1", LastAdded: 1},
				"2": &MockedEntry{ID: "2", Data: "2", LastAdded: 3},
			},
		}
		remote := &MockedEntryList{
			MockedEntries: []*MockedEntry{
				&MockedEntry{ID: "1", Data: "9", LastAdded: 4},
				&MockedEntry{ID: "3", Data: "1", LastAdded: 6},
			},
		}
		delta := store.ComputeDelta(remote)

		require.Equal(t, 2, delta.Length())
		entry := delta.AtIndex(0)
		require.NotNil(t, entry)
		mockedEntry := entry.(*MockedEntry)
		require.Equal(t, "9", mockedEntry.Data)
		require.Equal(t, "1", mockedEntry.ID)

		entry = delta.AtIndex(1)
		require.NotNil(t, entry)
		mockedEntry = entry.(*MockedEntry)
		require.Equal(t, "1", mockedEntry.Data)
		require.Equal(t, "3", mockedEntry.ID)
		store.ApplyDelta(delta)
		entry, err := store.backend.EntryByID("1")
		require.NoError(t, err)
		mockedEntry = entry.(*MockedEntry)
		require.Equal(t, "9", mockedEntry.Data)
		require.Equal(t, "1", mockedEntry.ID)

		entry, err = store.backend.EntryByID("3")
		require.NoError(t, err)
		mockedEntry = entry.(*MockedEntry)
		require.Equal(t, "1", mockedEntry.Data)
		require.Equal(t, "3", mockedEntry.ID)
	})
}
