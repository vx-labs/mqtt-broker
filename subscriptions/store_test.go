package subscriptions

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/mesh"
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

func TestStore(t *testing.T) {
	s, err := NewMemDBStore(&mockRouter{})
	require.NoError(t, err)
	err = s.Create(&Subscription{
		ID:        "1",
		Tenant:    "_default",
		Qos:       1,
		SessionID: "1",
		Pattern:   []byte("devices/+/degrees"),
	})
	require.NoError(t, err)
	sub, err := s.ByID("1")
	require.NoError(t, err)
	require.Equal(t, sub.SessionID, "1")
	set, err := s.ByTopic("_default", []byte("devices/a/degrees"))
	require.NoError(t, err)
	require.Equal(t, 1, len(set.Subscriptions))
	require.Equal(t, "1", set.Subscriptions[0].SessionID)
}
