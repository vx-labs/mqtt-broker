package broker

import (
	"log"

	publishQueue "github.com/vx-labs/mqtt-broker/queues/publish"
	"github.com/vx-labs/mqtt-broker/topics"
)

func (b *Broker) startPublishConsumers() {
	for i := 0; i < 5; i++ {
		go b.publishQueue.Consume(func(p *publishQueue.Message) {
			err := b.consumePublish(p)
			if err != nil {
				log.Printf("ERR: failed to publish message: %v", err)
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
			log.Printf("WARN: failed to save retained message: %v", err)
		}
	}
	err := b.routeMessage(message.Tenant, p)
	if err != nil {
		log.Printf("ERR: failed to route message: %v", err)
		return err
	}
	return nil
}
