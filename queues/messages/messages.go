package messages

import (
	"errors"
	"sync"
	"time"

	"github.com/google/btree"

	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrQueueEmpty = errors.New("queue empty")
	ErrQueueFull  = errors.New("queue is full")
)

type Queue struct {
	mutex    sync.Mutex
	elements *btree.BTree
}
type item struct {
	timestamp time.Time
	publish   *packet.Publish
}

func (i item) Less(remote btree.Item) bool {
	return i.timestamp.Before(remote.(item).timestamp)
}

func NewQueue() *Queue {
	return &Queue{
		elements: btree.New(5),
	}
}

func (q *Queue) Close() error {
	q.elements.Clear(true)
	return nil
}
func (q *Queue) Put(timestamp time.Time, p *packet.Publish) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// TODO: error handling
	if q.elements.Len() >= 5000 {
		return ErrQueueFull
	}
	q.elements.ReplaceOrInsert(item{
		publish:   p,
		timestamp: timestamp,
	})
	return nil
}
func (q *Queue) Pop() (*packet.Publish, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	elt := q.elements.DeleteMin()
	if elt == nil {
		return nil, ErrQueueEmpty
	}
	return elt.(item).publish, nil
}
