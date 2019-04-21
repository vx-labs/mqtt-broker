package subscriptions

import (
	"testing"

	"github.com/vx-labs/mqtt-broker/broker/cluster"

	"github.com/stretchr/testify/require"
)

func returnNilErr() error {
	return nil
}

func TestStore(t *testing.T) {
	s, err := NewMemDBStore(cluster.MockedMesh())
	require.NoError(t, err)
	err = s.Create(Subscription{Metadata: Metadata{
		ID:        "1",
		Tenant:    "_default",
		Peer:      "1",
		Qos:       1,
		SessionID: "1",
		Pattern:   []byte("devices/+/degrees"),
	}}, returnNilErr)
	require.NoError(t, err)
	sub, err := s.ByID("1")
	require.NoError(t, err)
	require.Equal(t, sub.SessionID, "1")
	set, err := s.ByTopic("_default", []byte("devices/a/degrees"))
	require.NoError(t, err)
	require.Equal(t, 1, len(set))
	require.Equal(t, "1", set[0].SessionID)
}
