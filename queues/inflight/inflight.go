package inflight

import (
	"errors"
	"time"

	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrQueueFull = errors.New("queue is full")
)

type Queue struct {
	messages        chan *packet.Publish
	acknowledgement chan int32
	sender          func(*packet.Publish) error
	stop            chan struct{}
}

func (q *Queue) Close() error {
	close(q.stop)
	return nil
}

func (q *Queue) retryDeliver(publish *packet.Publish) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		err := q.sender(publish)
		if err != nil {
			continue
		}
		if publish.Header.Qos == 0 {
			return
		}
		select {
		case ack := <-q.acknowledgement:
			if publish.MessageId == ack {
				return
			}
		case <-ticker.C:
			continue
		}
	}
}
func New(sender func(*packet.Publish) error) *Queue {
	q := &Queue{
		stop:            make(chan struct{}),
		messages:        make(chan *packet.Publish),
		acknowledgement: make(chan int32),
		sender:          sender,
	}
	go func() {
		for {
			select {
			case <-q.stop:
				return
			case publish := <-q.messages:
				q.retryDeliver(publish)
			}
		}
	}()
	return q
}

func (q *Queue) Put(publish *packet.Publish) {
	publish.MessageId = 1
	q.messages <- publish
}
func (q *Queue) Ack(i int32) {
	q.acknowledgement <- i
}
