package sessions

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vx-labs/mqtt-broker/broker/cluster"
)

const (
	sessionID = "cb8f3900-4146-4499-a880-c01611a6d9ee"
)

func returnNilErr() error {
	return nil
}
func TestSessionStore(t *testing.T) {
	store, _ := NewSessionStore(cluster.MockedMesh(), func(string, string) Transport {
		return nil
	}, log.New(ioutil.Discard, "", 0))

	t.Run("create", func(t *testing.T) {
		err := store.Upsert(Session{
			Metadata: Metadata{
				ID:       sessionID,
				ClientID: "test1",
				Peer:     "1",
			},
		}, nil)
		require.Nil(t, err)
		err = store.Upsert(Session{
			Metadata: Metadata{
				ID:       "3",
				ClientID: "test2",
				Peer:     "2",
			},
		}, nil)
		require.Nil(t, err)
	})

	t.Run("lookup", lookup(store, sessionID))
	t.Run("All", func(t *testing.T) {
		set, err := store.All()
		require.Nil(t, err)
		require.Equal(t, 2, len(set))
	})
	t.Run("lookup peer", func(t *testing.T) {
		set, err := store.ByPeer("2")
		require.Nil(t, err)
		require.Equal(t, 1, len(set))
		require.Equal(t, "3", set[0].ID)
	})
	t.Run("lookup client id", func(t *testing.T) {
		sessions, err := store.ByClientID("test1")
		require.Nil(t, err)
		require.Equal(t, 1, len(sessions))
		session := sessions[0]
		require.NotNil(t, session)
		require.Equal(t, sessionID, session.ID)
	})

	t.Run("delete", func(t *testing.T) {
		_, err := store.ByID(sessionID)
		require.NoError(t, err)
		err = store.Delete(sessionID, "test")
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
