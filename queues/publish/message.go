package publish

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/vx-labs/mqtt-protocol/packet"
)

type Message struct {
	Tenant  string
	Publish *packet.Publish
}

type node struct {
	item *Message
	next *inode
}
type inode struct {
	value *node
}

func (i *inode) get() *node {
	return i.value
}

func (i *inode) cas(old, new *node) bool {
	oldPtr := (*unsafe.Pointer)(unsafe.Pointer(&i.value))
	return atomic.CompareAndSwapPointer(oldPtr, unsafe.Pointer(old), unsafe.Pointer(new))
}

type queue struct {
	head *inode
	tail *inode
}

func New() *queue {
	n := &node{
		next: &inode{},
	}
	queue := &queue{
		head: &inode{
			value: n,
		},
		tail: &inode{
			value: n,
		},
	}
	return queue
}

func (q *queue) Enqueue(p *Message) {
	newNode := &node{
		item: p,
		next: &inode{},
	}
	for {
		currentTail := q.tail.get()
		tailNext := currentTail.next.get()

		if currentTail == q.tail.get() {
			if tailNext != nil {
				q.tail.cas(currentTail, tailNext)
			} else {
				if currentTail.next.cas(nil, newNode) {
					q.tail.cas(currentTail, newNode)
					return
				}
			}
		}
	}
}
func (q *queue) Pop() *Message {
	for {
		first := q.head.get()
		last := q.tail.get()
		next := first.next.get()
		if first == q.head.get() {
			if first == last {
				if next == nil {
					return nil
				}
				q.tail.cas(last, next)
			} else {
				item := next.item
				if q.head.cas(first, next) {
					return item
				}
			}
		}
	}
}

func (q *queue) Consume(f func(*Message)) {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		publish := q.Pop()
		if publish != nil {
			f(publish)
		} else {
			<-ticker.C
		}
	}
}
