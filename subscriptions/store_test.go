package subscriptions

import (
	"context"
	"testing"

	"github.com/vx-labs/mqtt-broker/subscriptions/pb"
	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/stretchr/testify/require"
)

func returnNilErr(context.Context, packet.Publish) error {
	return nil
}
func TestStore(t *testing.T) {
	s := NewMemDBStore()
	err := s.Create(&pb.Metadata{
		ID:        "1",
		Tenant:    "_default",
		Peer:      "1",
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
	require.Equal(t, 1, len(set.Metadatas))
	require.Equal(t, "1", set.Metadatas[0].SessionID)
}
