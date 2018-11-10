package inflight

import (
	"time"

	"github.com/vx-labs/mqtt-protocol/packet"
)

type Message struct {
	ID               int32
	Publish          *packet.Publish
	inflight         uint32
	consumed         uint32
	inflightDeadline time.Time
}

type MessageList []*Message

func (m MessageList) Apply(f func(*Message)) {
	for _, message := range m {
		f(message)
	}
}
func (m MessageList) ApplyE(f func(*Message) error) (err error) {
	for _, message := range m {
		err = f(message)
		if err != nil {
			return err
		}
	}
	return nil
}
func (m MessageList) Find(f func(*Message) bool) *Message {
	for _, message := range m {
		if f(message) {
			return message
		}
	}
	return nil
}
func (m MessageList) Filter(f func(*Message) bool) MessageList {
	copy := make(MessageList, 0)
	for _, message := range m {
		if f(message) {
			copy = append(copy, message)
		}
	}
	return copy
}
