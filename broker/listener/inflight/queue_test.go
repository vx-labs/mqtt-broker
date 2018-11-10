package inflight

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vx-labs/mqtt-protocol/packet"
)

func TestQueue(t *testing.T) {
	t.Run("max size", func(t *testing.T) {
		a := New(2)
		require.NoError(t, a.Insert(&packet.Publish{}))
		require.NoError(t, a.Insert(&packet.Publish{}))
		require.Error(t, a.Insert(&packet.Publish{}))
	})
	t.Run("find free ID", func(t *testing.T) {
		a := New(3)
		require.NoError(t, a.Insert(&packet.Publish{}))
		require.NoError(t, a.Insert(&packet.Publish{}))
		a.messages[0].consumed = 0
		a.messages[0].inflight = 1
		a.messages[0].ID = 1

		a.messages[1].consumed = 0
		a.messages[1].inflight = 1
		a.messages[1].ID = 3
		require.Equal(t, int32(2), a.freeInflightID())
	})

	t.Run("next", func(t *testing.T) {
		a := New(2)
		require.NoError(t, a.Insert(&packet.Publish{}))
		p := a.Next()
		require.NotNil(t, p)
	})
	t.Run("expire inflight", func(t *testing.T) {
		a := New(2)
		a.messages = append(a.messages, &Message{
			inflight:         1,
			ID:               1,
			inflightDeadline: time.Time{},
			Publish:          &packet.Publish{Header: &packet.Header{}},
		},
			&Message{
				inflight:         1,
				ID:               2,
				inflightDeadline: time.Now().Add(10 * time.Second),
				Publish:          &packet.Publish{Header: &packet.Header{}},
			})
		require.Equal(t, 2, len(a.messages))
		require.Equal(t, 2, cap(a.messages))
		require.Equal(t, uint32(1), a.messages[0].inflight)
		require.Equal(t, uint32(1), a.messages[1].inflight)
		a.ExpireInflight()
		require.Equal(t, 2, len(a.messages))
		require.Equal(t, 2, cap(a.messages))
		require.Equal(t, uint32(0), a.messages[0].inflight)
		require.Equal(t, uint32(1), a.messages[1].inflight)
	})
	t.Run("inflight id", func(t *testing.T) {
		a := New(3)
		a.Insert(&packet.Publish{})
		a.Insert(&packet.Publish{})
		a.Insert(&packet.Publish{})

		m := a.Next()
		require.Equal(t, int32(1), m.ID)
		m = a.Next()
		require.Equal(t, int32(2), m.ID)
		m = a.Next()
		require.Equal(t, int32(3), m.ID)
	})
}
