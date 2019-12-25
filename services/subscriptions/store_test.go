package subscriptions

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/services/subscriptions/pb"
	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/stretchr/testify/require"
)

func returnNilErr(context.Context, packet.Publish) error {
	return nil
}
func TestStore(t *testing.T) {
	s := NewSubscriptionStore(cluster.MockedMesh(), zap.NewNop())
	err := s.Create(&pb.Subscription{
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
	require.Equal(t, 1, len(set.Subscriptions))
	require.Equal(t, "1", set.Subscriptions[0].SessionID)

}
