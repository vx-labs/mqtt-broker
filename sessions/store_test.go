package sessions

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const (
	sessionID = "cb8f3900-4146-4499-a880-c01611a6d9ee"
)

func TestSessionStore(t *testing.T) {
	store := NewSessionStore()

	t.Run("create", func(t *testing.T) {
		err := store.Upsert(&Session{
			ID:   sessionID,
			Peer: 1,
		})
		assert.Nil(t, err)
		err = store.Upsert(&Session{
			ID:   "3",
			Peer: 2,
		})
		assert.Nil(t, err)
	})

	t.Run("lookup", lookup(store, sessionID))
	t.Run("All", func(t *testing.T) {
		set, err := store.All()
		assert.Nil(t, err)
		assert.Equal(t, 2, len(set.Sessions))
	})
	t.Run("lookup peer", func(t *testing.T) {
		set, err := store.ByPeer(2)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(set.Sessions))
		assert.Equal(t, "3", set.Sessions[0].ID)
	})

	t.Run("delete", func(t *testing.T) {
		err := store.Delete(sessionID)
		assert.Nil(t, err)
		_, err = store.ById(sessionID)
		assert.NotNil(t, err)
	})
	t.Run("merge", func(t *testing.T) {
		remote := NewSessionStore()
		require.NoError(t, remote.Upsert(&Session{
			ID:   "a",
			Peer: 5,
		}))
		delta := store.ComputeDelta(remote.DumpState())
		require.Equal(t, 1, len(delta.Sessions))
		require.Equal(t, "a", delta.Sessions[0].ID)

		require.NoError(t, remote.Upsert(&Session{
			ID:   "3",
			Peer: 5,
		}))
		delta = store.ComputeDelta(remote.DumpState())
		require.Equal(t, 2, len(delta.Sessions))
		require.Equal(t, "a", delta.Sessions[1].ID)
		require.Equal(t, "3", delta.Sessions[0].ID)
	})
}

func lookup(store SessionStore, id string) func(*testing.T) {
	return func(t *testing.T) {
		sess, err := store.ById(id)
		assert.Nil(t, err)
		assert.Equal(t, id, sess.ID)
	}
}
