package inflight

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrNotFound        = fmt.Errorf("queue not found")
	ErrMessageNotFound = fmt.Errorf("message not found")
)

type Queue struct {
	depth    int
	messages MessageList
	mutex    sync.Mutex
	notify   chan struct{}
	quit     chan struct{}
}

func New(depth int) *Queue {
	return &Queue{
		depth:    depth,
		messages: make(MessageList, 0, depth),
		notify:   make(chan struct{}),
		quit:     make(chan struct{}),
	}
}
func (q *Queue) Capacity() int {
	return cap(q.messages)
}
func (q *Queue) Empty() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.messages = make(MessageList, 0, q.depth)
}
func (q *Queue) MessageCount() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.messages.Filter(func(m *Message) bool {
		return m.consumed == 0 &&
			m.inflight == 0
	}))
}
func (q *Queue) freeInflightID() int32 {
	inflightMessages := q.messages.Filter(func(m *Message) bool {
		return m.consumed == 0 &&
			m.inflight == 1
	})
	if len(inflightMessages) == 0 {
		return 1
	}
	flags := make([]bool, q.depth)
	inflightMessages.Apply(func(m *Message) {
		flags[m.ID-1] = true
	})
	for idx, f := range flags {
		if !f {
			return int32(idx + 1)
		}
	}
	return int32(q.depth)
}
func (q *Queue) InflightCount() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.messages.Filter(func(m *Message) bool {
		return m.consumed == 0 &&
			m.inflight == 1
	}))
}

func (q *Queue) ExpireInflight() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	now := time.Now()
	set := q.messages.Filter(func(m *Message) bool {
		return m.consumed == 0 &&
			m.inflight == 1 &&
			now.After(m.inflightDeadline)
	})
	if len(set) == 0 {
		return
	}
	set.Apply(func(message *Message) {
		message.inflight = 0
		message.inflightDeadline = time.Time{}
		message.ID = 0
		message.Publish.MessageId = 0
		message.Publish.Header.Dup = true
	})
	q.sendNotify()
}
func (q *Queue) Close() error {
	q.messages = nil
	close(q.quit)
	return nil
}
func (q *Queue) Next() *Message {
	q.mutex.Lock()
	notify := q.notify
	p := q.messages.Find(func(m *Message) bool {
		return m.inflight == 0 && m.consumed == 0
	})
	if p == nil {
		q.mutex.Unlock()
		select {
		case <-notify:
		case <-q.quit:
			close(q.notify)
			return nil
		}
		return q.Next()
	}
	defer q.mutex.Unlock()
	p.ID = q.freeInflightID()
	p.inflight = 1
	p.inflightDeadline = time.Now().Add(10 * time.Second)
	p.Publish.MessageId = p.ID
	return p
}
func (q *Queue) sendNotify() {
	ch := q.notify
	q.notify = make(chan struct{})
	close(ch)
}
func (q *Queue) Insert(p *packet.Publish) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.insert(p)
}
func (q *Queue) insert(p *packet.Publish) error {
	if len(q.messages) == cap(q.messages) {
		return errors.New("queue is full")
	}
	q.messages = append(q.messages, &Message{
		consumed: 0,
		inflight: 0,
		Publish:  p,
		ID:       0,
	})
	q.sendNotify()
	return nil
}
func (q *Queue) Acknowledge(messageID int32) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	message := q.messages.Find(func(message *Message) bool {
		return message.ID == messageID
	})
	if message == nil {
		return ErrMessageNotFound
	}
	message.inflight = 0
	message.consumed = 1
	message.ID = 0
	message.Publish.MessageId = 0
	list := q.messages.Filter(func(m *Message) bool {
		return m.consumed == 0
	})
	q.messages = make(MessageList, len(list), q.depth)
	copy(q.messages, list)
	return nil
}
