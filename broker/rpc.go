package broker

import (
	"github.com/vx-labs/mqtt-broker/topics"
)

func (b *Broker) ListRetainedMessages() (topics.RetainedMessageSet, error) {
	return b.Topics.All()
}
