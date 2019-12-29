package sessions

import (
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
	"github.com/vx-labs/mqtt-broker/services/sessions/pb"
)

const (
	sessionID = "cb8f3900-4146-4499-a880-c01611a6d9ee"
)

func returnNilErr() error {
	return nil
}
func TestSessionStore(t *testing.T) {
	store := NewSessionStore(zap.NewNop())

	t.Run("create", func(t *testing.T) {
		err := store.Create(&pb.Session{
			ID:       sessionID,
			Tenant:   "_default",
			ClientID: "test1",
			Peer:     "1",
		})
		require.Nil(t, err)
		err = store.Create(&pb.Session{
			ID:       "3",
			Tenant:   "_default",
			ClientID: "test2",
			Peer:     "2",
		})
		require.Nil(t, err)
	})

	t.Run("lookup", lookup(store, sessionID))
	t.Run("All", func(t *testing.T) {
		set, err := store.All(nil)
		require.Nil(t, err)
		require.Equal(t, 2, len(set.Sessions))
	})
	t.Run("All ID filtered", func(t *testing.T) {
		set, err := store.All(&pb.SessionFilterInput{ID: []string{"3"}})
		require.Nil(t, err)
		require.Equal(t, 1, len(set.Sessions))
		require.Equal(t, "3", set.Sessions[0].ID)
	})
	t.Run("lookup peer", func(t *testing.T) {
		set, err := store.ByPeer("2")
		require.Nil(t, err)
		require.Equal(t, 1, len(set.Sessions))
		require.Equal(t, "3", set.Sessions[0].ID)
	})
	t.Run("lookup client id", func(t *testing.T) {
		sessions, err := store.ByClientID("test1")
		require.Nil(t, err)
		require.Equal(t, 1, len(sessions.Sessions))
		session := sessions.Sessions[0]
		require.NotNil(t, session)
		require.Equal(t, sessionID, session.ID)
	})

	t.Run("delete", func(t *testing.T) {
		_, err := store.ByID(sessionID)
		require.NoError(t, err)
		err = store.Delete(sessionID)
		require.NoError(t, err)
		_, err = store.ByID(sessionID)
		require.Error(t, err)
	})
}

func lookup(store SessionStore, id string) func(*testing.T) {
	return func(t *testing.T) {
		sess, err := store.ByID(id)
		require.Nil(t, err)
		require.Equal(t, id, sess.ID)
	}
}
