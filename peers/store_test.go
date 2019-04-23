package peers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vx-labs/mqtt-broker/broker/cluster"
)

const (
	peerID = "cb8f3900-4146-4499-a880-c01611a6d9ee"
)

func TestPeerStore(t *testing.T) {
	store, _ := NewPeerStore(cluster.MockedMesh())

	t.Run("create", func(t *testing.T) {
		err := store.Upsert(Peer{Metadata: Metadata{
			ID:     peerID,
			MeshID: "1",
		}})
		assert.Nil(t, err)
		err = store.Upsert(Peer{Metadata: Metadata{
			ID:     "3",
			MeshID: "2",
		}})
		assert.Nil(t, err)
	})

	t.Run("lookup", lookup(store, peerID))
	t.Run("All", func(t *testing.T) {
		set, err := store.All()
		require.Nil(t, err)
		assert.Equal(t, 2, len(set))
	})
	t.Run("lookup peer", func(t *testing.T) {
		set, err := store.ByMeshID("2")
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
