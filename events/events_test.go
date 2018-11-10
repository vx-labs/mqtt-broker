package events

import (
	"testing"
)

func TestEvents(t *testing.T) {
	bus := NewEventBus()
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

func BenchmarkEvents_Emit(b *testing.B) {
	bus := NewEventBus()
	cancel := bus.Subscribe("entry_added", func(_ Event) {})
	defer cancel()

	for i := 0; i < b.N; i++ {
		bus.Emit(Event{
			Key:   "entry_added",
			Entry: nil,
		})
	}
}
func BenchmarkEvents_Subscribe(b *testing.B) {
	bus := NewEventBus()
	for i := 0; i < b.N; i++ {
		cancel := bus.Subscribe("entry_added", func(_ Event) {})
		cancel()
	}
}
