package peers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
)

const (
	peerID = "cb8f3900-4146-4499-a880-c01611a6d9ee"
)

func TestPeerStore(t *testing.T) {
	store, _ := NewPeerStore()

	t.Run("create", func(t *testing.T) {
		err := store.Upsert(&pb.Peer{
			ID: peerID,
		})
		require.Nil(t, err)
		err = store.Upsert(&pb.Peer{
			ID: "3",
		})
		require.Nil(t, err)
	})

	t.Run("lookup", lookup(store, peerID))
	t.Run("All", func(t *testing.T) {
		set, err := store.All()
		require.Nil(t, err)
		require.Equal(t, 2, len(set.Peers))
	})
	t.Run("delete", func(t *testing.T) {
		err := store.Delete(peerID)
		require.NoError(t, err)
		_, err = store.ByID(peerID)
		require.Error(t, err)
	})
}

func lookup(store PeerStore, id string) func(*testing.T) {
	return func(t *testing.T) {
		sess, err := store.ByID(id)
		require.Nil(t, err)
		assert.Equal(t, id, sess.ID)
	}
}
