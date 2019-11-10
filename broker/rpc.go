package broker

import topicspb "github.com/vx-labs/mqtt-broker/topics/pb"

func (b *Broker) ListRetainedMessages() (topicspb.RetainedMessageSet, error) {
	return b.Topics.All()
}
