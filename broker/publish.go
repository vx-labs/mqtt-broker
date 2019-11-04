package broker

import (
	queues "github.com/vx-labs/mqtt-broker/queues/pb"
	publishQueue "github.com/vx-labs/mqtt-broker/struct/queues/publish"
	"github.com/vx-labs/mqtt-broker/topics"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func (b *Broker) startPublishConsumers() {
	jobs := make(chan chan *publishQueue.Message)
	for i := 0; i < 5; i++ {
		go func() {
			ch := make(chan *publishQueue.Message)
			defer close(ch)
			for {
				// TODO: implement termination
				jobs <- ch
				p := <-ch
				err := b.consumePublish(p)
				if err != nil {
					b.logger.Error("failed to publish message", zap.Binary("topic_pattern", p.Publish.Topic), zap.Error(err))
					//	b.publishQueue.Enqueue(p)
				}
			}
		}()
	}
	go b.publishQueue.Consume(func(message *publishQueue.Message) {
		ch := <-jobs
		ch <- message
	})
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
			b.logger.Error("failed to save retained message", zap.Binary("topic_pattern", message.Topic), zap.Error(err))
		}
	}
	err := b.routeMessage(message.Tenant, p)
	if err != nil {
		b.logger.Error("failed to route message", zap.Binary("topic_pattern", message.Publish.Topic), zap.Error(err))
		return err
	}
	return nil
}

type batchedMessage struct {
	Peer       string
	Recipients []string
}

func (b *Broker) routeMessage(tenant string, p *packet.Publish) error {
	recipients, err := b.Subscriptions.ByTopic(b.ctx, tenant, p.Topic)
	if err != nil {
		return err
	}
	message := *p
	message.Header.Retain = false
	payload := make([]queues.MessageBatch, len(recipients))
	for idx := range recipients {
		payload[idx] = queues.MessageBatch{ID: recipients[idx].SessionID, Publish: &message}
	}
	err = b.Queues.PutMessageBatch(b.ctx, payload)
	if err != nil {
		b.logger.Warn("failed to enqueue message", zap.Error(err))
	}
	return nil
}
