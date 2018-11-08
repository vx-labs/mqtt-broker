package events

import (
	"testing"

	"github.com/hashicorp/go-immutable-radix"
)

func TestEvents(t *testing.T) {
	bus := EventBus{
		state: iradix.New(),
	}
	events, cancel := bus.Subscribe()

	done := make(chan struct{})
	go func() {
		<-events
		close(done)
	}()
	bus.Emit(Event{
		Key:   "entry_added",
		Entry: nil,
	})
	<-done
	cancel()
}

func BenchmarkEvents(b *testing.B) {
	bus := EventBus{
		state: iradix.New(),
	}
	events, cancel := bus.Subscribe()
	defer cancel()
	go func() {
		for range events {
		}
	}()
	for i := 0; i < b.N; i++ {
		bus.Emit(Event{
			Key:   "entry_added",
			Entry: nil,
		})
	}
}
