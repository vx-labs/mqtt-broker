package events

import (
	"testing"

	"github.com/hashicorp/go-immutable-radix"
)

func TestEvents(t *testing.T) {
	bus := Bus{
		state: iradix.New(),
	}
	events, cancel := bus.Subscribe("entry_added")

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
	bus := Bus{
		state: iradix.New(),
	}
	events, cancel := bus.Subscribe("entry_added")
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
