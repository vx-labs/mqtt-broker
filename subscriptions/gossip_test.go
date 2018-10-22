package subscriptions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionList(t *testing.T) {
	local := &SubscriptionList{
		Subscriptions: []*Subscription{
			&Subscription{
				ID:          "1",
				LastUpdated: 1,
			},
			&Subscription{
				ID:          "2",
				Peer:        1,
				LastUpdated: 1,
			},
		},
	}
	remote := &SubscriptionList{
		Subscriptions: []*Subscription{
			&Subscription{
				ID:          "2",
				Peer:        2,
				LastUpdated: 2,
			},
		},
	}
	data := local.Merge(remote)
	delta := data.(*SubscriptionList)
	require.Equal(t, 1, len(delta.Subscriptions))
	require.Equal(t, uint64(2), delta.Subscriptions[0].Peer)
	require.Equal(t, uint64(2), local.Subscriptions[1].Peer)
}
