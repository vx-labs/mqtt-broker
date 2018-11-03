package sessions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionList(t *testing.T) {
	local := &SessionList{
		Sessions: []*Session{
			&Session{
				ID:          "1",
				LastUpdated: 1,
			},
			&Session{
				ID:          "2",
				Peer:        1,
				LastUpdated: 1,
			},
		},
	}
	remote := &SessionList{
		Sessions: []*Session{
			&Session{
				ID:          "2",
				Peer:        2,
				LastUpdated: 2,
			},
		},
	}
	data := local.Merge(remote)
	delta := data.(*SessionList)
	require.Equal(t, 1, len(delta.Sessions))
	require.Equal(t, uint64(2), delta.Sessions[0].Peer)
	require.Equal(t, uint64(2), local.Sessions[1].Peer)
}
