package inflight

import (
	"context"
	"errors"
	"log"

	"github.com/google/btree"

	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrQueueFull = errors.New("queue is full")
)

type Job struct {
	publish *packet.Publish
	onAck   func()
}
type Queue struct {
	acknowlegers *btree.BTree
	messages     chan *packet.Publish
	sender       func(*packet.Publish) error
	stop         chan struct{}
	jobs         chan chan Job
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
		jobs:         make(chan chan Job),
	}
	var i int32
	for i = 1; i <= 10; i++ {
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

func (q *Queue) Put(ctx context.Context, publish *packet.Publish, onAck func()) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ack := <-q.jobs:
		select {
		case ack <- Job{
			onAck:   onAck,
			publish: publish,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
func (q *Queue) Ack(p *packet.PubAck) error {
	ack := q.acknowlegers.Get(&acknowleger{
		mid: p.MessageId,
	})
	if ack == nil {
		log.Printf("WARN: received ack for an unknown message id: %d", p.MessageId)
		return errors.New("id not found")
	}
	ack.(*acknowleger).ch <- p
	return nil
}
