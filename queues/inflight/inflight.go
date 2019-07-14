package inflight

import (
	"errors"
	"log"

	"github.com/google/btree"

	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrQueueFull = errors.New("queue is full")
)

type Queue struct {
	acknowlegers *btree.BTree
	messages     chan *packet.Publish
	sender       func(*packet.Publish) error
	stop         chan struct{}
	jobs         chan chan *packet.Publish
}

type acknowleger struct {
	mid    int32
	ch     chan<- *packet.PubAck
	cancel chan struct{}
	quit   chan struct{}
}

func (a *acknowleger) Less(remote btree.Item) bool {
	return a.mid < remote.(*acknowleger).mid
}

func (q *Queue) Close() error {
	close(q.stop)
	return nil
}

func New(sender func(*packet.Publish) error) *Queue {
	q := &Queue{
		stop:         make(chan struct{}),
		messages:     make(chan *packet.Publish),
		sender:       sender,
		acknowlegers: btree.New(2),
		jobs:         make(chan chan *packet.Publish),
	}
	var i int32
	for i = 0; i < 10; i++ {
		q.acknowlegers.ReplaceOrInsert(startDeliverer(i, q.jobs, sender))
	}
	go func() {
		for {
			select {
			case <-q.stop:
				q.acknowlegers.Descend(func(i btree.Item) bool {
					ack := i.(*acknowleger)
					close(ack.cancel)
					<-ack.quit
					return true
				})
				return
			}
		}
	}()
	return q
}

func (q *Queue) Put(publish *packet.Publish) {
	ack := <-q.jobs
	ack <- publish
}
func (q *Queue) Ack(p *packet.PubAck) {
	ack := q.acknowlegers.Get(&acknowleger{
		mid: p.MessageId,
	})
	if ack == nil {
		log.Printf("WARN: received ack for an unknown message id: %d", p.MessageId)
		return
	}
	ack.(*acknowleger).ch <- p
}
