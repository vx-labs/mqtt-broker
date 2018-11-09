package events

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/google/uuid"
	"github.com/hashicorp/go-immutable-radix"
)

type Event struct {
	Key   string
	Entry interface{}
}

type subscription struct {
	handler func(Event)
}

func newSubscription(handler func(Event)) *subscription {
	return &subscription{
		handler: handler,
	}
}

func (sub *subscription) send(ev Event) {
	sub.handler(ev)
}
func (sub *subscription) close() {
}

type CancelFunc func()

type Bus struct {
	state *iradix.Tree
}

func (e *Bus) cas(old, new *iradix.Tree) bool {
	oldPtr := (*unsafe.Pointer)(unsafe.Pointer(&e.state))
	return atomic.CompareAndSwapPointer(oldPtr, unsafe.Pointer(old), unsafe.Pointer(new))
}

func (e *Bus) Emit(ev Event) {
	e.state.Root().WalkPrefix([]byte(ev.Key+"/"), func(k []byte, v interface{}) bool {
		sub := v.(*subscription)
		sub.send(ev)
		return false
	})
}
func (e *Bus) Subscribe(key string, handler func(Event)) func() {
	sub := newSubscription(handler)
	id := fmt.Sprintf("%s/%s", key, uuid.New().String())
	cancel := func() {
		for {
			old := e.state
			new, _, _ := old.Delete([]byte(id))
			if e.cas(old, new) {
				sub.close()
				return
			}
		}
	}
	for {
		old := e.state
		new, _, _ := old.Insert([]byte(id), sub)
		if e.cas(old, new) {
			return cancel
		}
	}
}

func NewEventBus() *Bus {
	return &Bus{
		state: iradix.New(),
	}
}
