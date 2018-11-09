package events

import (
	"testing"

	"github.com/hashicorp/go-immutable-radix"
)

func TestEvents(t *testing.T) {
	bus := Bus{
		state: iradix.New(),
	}
	done := make(chan struct{})

	cancel := bus.Subscribe("entry_added", func(_ Event) {
		close(done)
	})

	bus.Emit(Event{
		Key:   "entry_added",
		Entry: nil,
	})
	<-done
	cancel()
}

func BenchmarkEvents(b *testing.B) {
	bus := Bus{
		state: iradix.New(),
	}
	cancel := bus.Subscribe("entry_added", func(_ Event) {})
	defer cancel()
	for i := 0; i < b.N; i++ {
		bus.Emit(Event{
			Key:   "entry_added",
			Entry: nil,
		})
	}
}
