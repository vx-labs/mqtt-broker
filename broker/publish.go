package broker

import (
	publishQueue "github.com/vx-labs/mqtt-broker/queues/publish"
	"github.com/vx-labs/mqtt-broker/topics"
	"go.uber.org/zap"
)

func (b *Broker) startPublishConsumers() {
	for i := 0; i < 5; i++ {
		go b.publishQueue.Consume(func(p *publishQueue.Message) {
			err := b.consumePublish(p)
			if err != nil {
				b.logger.Error("failed to publish message", b.zapNodeID(), zap.Binary("topic_pattern", p.Publish.Topic), zap.Error(err))
				b.publishQueue.Enqueue(p)
			}
		})
	}
}
func (b *Broker) consumePublish(message *publishQueue.Message) error {
	p := message.Publish
	if p.Header.Retain {
		message := topics.RetainedMessage{
			Metadata: topics.Metadata{
				Payload: p.Payload,
				Qos:     p.Header.Qos,
				Tenant:  message.Tenant,
				Topic:   p.Topic,
			},
		}
		err := b.Topics.Create(message)
		if err != nil {
			b.logger.Error("failed to save retained message", b.zapNodeID(), zap.Binary("topic_pattern", message.Topic), zap.Error(err))
		}
	}
	err := b.routeMessage(message.Tenant, p)
	if err != nil {
		b.logger.Error("failed to route message", b.zapNodeID(), zap.Binary("topic_pattern", message.Publish.Topic), zap.Error(err))
		return err
	}
	return nil
}
