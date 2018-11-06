package state

import (
	"testing"

	"github.com/hashicorp/go-immutable-radix"
)

func TestEvents(t *testing.T) {
	bus := EventBus{
		state: iradix.New(),
	}
	events, cancel := bus.Events()

	done := make(chan struct{})
	go func() {
		<-events
		close(done)
	}()
	bus.Emit(Event{
		Kind:  EntryAdded,
		Entry: nil,
	})
	<-done
	cancel()
}

func BenchmarkEvents(b *testing.B) {
	bus := EventBus{
		state: iradix.New(),
	}
	events, cancel := bus.Events()
	defer cancel()
	go func() {
		for range events {
		}
	}()
	for i := 0; i < b.N; i++ {
		bus.Emit(Event{
			Kind:  EntryAdded,
			Entry: nil,
		})
	}
}
