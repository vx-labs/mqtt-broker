package messages

import (
	"errors"
	"time"

	"github.com/google/btree"

	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrQueueEmpty = errors.New("queue empty")
)

type Queue struct {
	elements *btree.BTree
}
type item struct {
	timestamp time.Time
	publish   packet.Publish
}

func (i item) Less(remote btree.Item) bool {
	return i.timestamp.Before(remote.(item).timestamp)
}

func NewQueue() *Queue {
	return &Queue{
		elements: btree.New(5),
	}
}

func (q *Queue) Put(timestamp time.Time, p packet.Publish) {
	q.elements.ReplaceOrInsert(item{
		publish:   p,
		timestamp: timestamp,
	})
}
func (q *Queue) Pop() (packet.Publish, error) {
	elt := q.elements.DeleteMin()
	if elt == nil {
		return packet.Publish{}, ErrQueueEmpty
	}
	return elt.(item).publish, nil
}
