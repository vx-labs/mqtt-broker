package sessions

import (
	"testing"

	"github.com/weaveworks/mesh"

	"github.com/stretchr/testify/require"
)

const (
	sessionID = "cb8f3900-4146-4499-a880-c01611a6d9ee"
)

type mockGossip struct{}

func (m *mockGossip) GossipUnicast(dst mesh.PeerName, msg []byte) error {
	return nil
}

func (m *mockGossip) GossipBroadcast(update mesh.GossipData) {
}

type mockRouter struct {
}

func (m *mockRouter) NewGossip(channel string, gossiper mesh.Gossiper) (mesh.Gossip, error) {
	return &mockGossip{}, nil
}

func TestSessionStore(t *testing.T) {
	store, _ := NewSessionStore(&mockRouter{})

	t.Run("create", func(t *testing.T) {
		err := store.Upsert(Session{
			ID:       sessionID,
			ClientID: "test1",
			Peer:     1,
		})
		require.Nil(t, err)
		err = store.Upsert(Session{
			ID:       "3",
			ClientID: "test2",
			Peer:     2,
		})
		require.Nil(t, err)
	})

	t.Run("lookup", lookup(store, sessionID))
	t.Run("All", func(t *testing.T) {
		set, err := store.All()
		require.Nil(t, err)
		require.Equal(t, 2, len(set.Sessions))
	})
	t.Run("lookup peer", func(t *testing.T) {
		set, err := store.ByPeer(2)
		require.Nil(t, err)
		require.Equal(t, 1, len(set.Sessions))
		require.Equal(t, "3", set.Sessions[0].ID)
	})
	t.Run("lookup client id", func(t *testing.T) {
		sessions, err := store.ByClientID("test1")
		require.Nil(t, err)
		session := sessions.Sessions[0]
		require.NotNil(t, session)
		require.Equal(t, sessionID, session.ID)
	})

	t.Run("delete", func(t *testing.T) {
		err := store.Delete(sessionID, "test")
		require.Nil(t, err)
		_, err = store.ByID(sessionID)
		require.NotNil(t, err)
	})
}

func lookup(store SessionStore, id string) func(*testing.T) {
	return func(t *testing.T) {
		sess, err := store.ByID(id)
		require.Nil(t, err)
		require.Equal(t, id, sess.ID)
	}
}
