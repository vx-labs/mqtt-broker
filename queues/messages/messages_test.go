package messages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vx-labs/mqtt-protocol/packet"
)

func TestQueue(t *testing.T) {
	q := NewQueue()
	q.Put(time.Now(), &packet.Publish{MessageId: 5})
	p, err := q.Pop()
	require.NoError(t, err)
	require.Equal(t, int32(5), p.MessageId)
}

func BenchmarkQueue(b *testing.B) {
	q := NewQueue()
	stamp := time.Now()
	p := &packet.Publish{MessageId: 5}
	for i := 0; i < b.N; i++ {
		q.Put(stamp, p)
	}
}
