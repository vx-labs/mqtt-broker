package crdt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntry(t *testing.T) {
	require.True(t, IsEntryAdded(&MockedEntry{
		LastAdded:   1,
		LastDeleted: 0,
	}))
	require.True(t, IsEntryRemoved(&MockedEntry{
		LastAdded:   1,
		LastDeleted: 2,
	}))
	t.Run("Should flag entry outdated", func(t *testing.T) {
		local := &MockedEntry{
			LastAdded:   2,
			LastDeleted: 3,
		}
		t.Run("if remote was updated later", func(t *testing.T) {
			remote := &MockedEntry{
				LastAdded:   4,
				LastDeleted: 0,
			}
			require.True(t, IsEntryOutdated(local, remote))
		})
		t.Run("if remote was removed later", func(t *testing.T) {
			remote := &MockedEntry{
				LastAdded:   2,
				LastDeleted: 5,
			}
			require.True(t, IsEntryOutdated(local, remote))
		})
	})
}
