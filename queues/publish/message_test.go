package publish

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vx-labs/mqtt-protocol/packet"
)

func TestQueue(t *testing.T) {
	t.Run("enqueue", func(t *testing.T) {
		queue := New()

		queue.Enqueue(&Message{Publish: &packet.Publish{MessageId: 1}})
	})
	t.Run("pop", func(t *testing.T) {
		queue := New()

		queue.Enqueue(&Message{Publish: &packet.Publish{MessageId: 1}})
		item := queue.Pop()
		require.NotNil(t, item)
		require.Equal(t, int32(1), item.Publish.MessageId)
	})
}
func BenchmarkQueue(b *testing.B) {

	queue := New()
	p := &Message{Publish: &packet.Publish{MessageId: 1}}

	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			queue.Enqueue(p)
		}
	})
}
