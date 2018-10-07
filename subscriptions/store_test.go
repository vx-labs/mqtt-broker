package subscriptions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	s, err := NewMemDBStore()
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
	require.Equal(t, 1, len(set))
	require.Equal(t, "1", set[0].SessionID)
}
