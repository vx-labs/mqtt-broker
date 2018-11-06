package state

import (
	"sync/atomic"
	"unsafe"

	"github.com/google/uuid"
	"github.com/hashicorp/go-immutable-radix"
)

type EventKind int

const (
	EntryAdded EventKind = iota
	EntryRemoved
)

type Event struct {
	Kind  EventKind
	Entry Entry
}

type subscription struct {
	ch   chan Event
	quit chan struct{}
}

type CancelFunc func()

type EventBus struct {
	state *iradix.Tree
}

func (e *EventBus) cas(old, new *iradix.Tree) bool {
	oldPtr := (*unsafe.Pointer)(unsafe.Pointer(&e.state))
	return atomic.CompareAndSwapPointer(oldPtr, unsafe.Pointer(old), unsafe.Pointer(new))
}
func (e *EventBus) Emit(ev Event) {
	e.state.Root().Walk(func(k []byte, v interface{}) bool {
		sub := v.(*subscription)
		select {
		case <-sub.quit:
		case sub.ch <- ev:
		}
		return false
	})
}
func (e *EventBus) Events() (chan Event, func()) {
	sub := &subscription{
		ch:   make(chan Event),
		quit: make(chan struct{}),
	}
	id := uuid.New().String()
	cancel := func() {
		for {
			old := e.state
			new, _, _ := old.Delete([]byte(id))
			if e.cas(old, new) {
				close(sub.quit)
				close(sub.ch)
				return
			}
		}
	}
	for {
		old := e.state
		new, _, _ := old.Insert([]byte(id), sub)
		if e.cas(old, new) {
			return sub.ch, cancel
		}
	}
}

func NewEventBus() *EventBus {
	return &EventBus{
		state: iradix.New(),
	}
}
