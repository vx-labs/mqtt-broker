package sessions

import (
	"testing"

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
		assert.Equal(t, 2, len(set))
	})
	t.Run("lookup peer", func(t *testing.T) {
		set, err := store.ByPeer(2)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(set))
		assert.Equal(t, "3", set[0].ID)
	})

	t.Run("delete", func(t *testing.T) {
		err := store.Delete(sessionID)
		assert.Nil(t, err)
		_, err = store.ById(sessionID)
		assert.NotNil(t, err)
	})

}

func lookup(store SessionStore, id string) func(*testing.T) {
	return func(t *testing.T) {
		sess, err := store.ById(id)
		assert.Nil(t, err)
		assert.Equal(t, id, sess.ID)
	}
}
