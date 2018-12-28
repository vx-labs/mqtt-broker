package peers

import (
	"testing"

	"github.com/weaveworks/mesh"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	peerID = "cb8f3900-4146-4499-a880-c01611a6d9ee"
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

func TestPeerStore(t *testing.T) {
	store, _ := NewPeerStore(&mockRouter{})

	t.Run("create", func(t *testing.T) {
		err := store.Upsert(&Peer{
			ID: peerID,
		})
		assert.Nil(t, err)
		err = store.Upsert(&Peer{
			ID:     "3",
			MeshID: 2,
		})
		assert.Nil(t, err)
	})

	t.Run("lookup", lookup(store, peerID))
	t.Run("All", func(t *testing.T) {
		set, err := store.All()
		require.Nil(t, err)
		assert.Equal(t, 2, len(set.Peers))
	})
	t.Run("lookup peer", func(t *testing.T) {
		set, err := store.ByMeshID(2)
		require.Nil(t, err)
		assert.Equal(t, "3", set.ID)
	})

	t.Run("delete", func(t *testing.T) {
		err := store.Delete(peerID)
		assert.Nil(t, err)
		_, err = store.ByID(peerID)
		assert.NotNil(t, err)
	})
}

func lookup(store PeerStore, id string) func(*testing.T) {
	return func(t *testing.T) {
		sess, err := store.ByID(id)
		require.Nil(t, err)
		assert.Equal(t, id, sess.ID)
	}
}
