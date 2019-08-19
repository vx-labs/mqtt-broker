package broker

import (
	listenerpb "github.com/vx-labs/mqtt-broker/listener/pb"
	publishQueue "github.com/vx-labs/mqtt-broker/queues/publish"
	"github.com/vx-labs/mqtt-broker/topics"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (b *Broker) startPublishConsumers() {
	for i := 0; i < 5; i++ {
		go b.publishQueue.Consume(func(p *publishQueue.Message) {
			err := b.consumePublish(p)
			if err != nil {
				b.logger.Error("failed to publish message", zap.Binary("topic_pattern", p.Publish.Topic), zap.Error(err))
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
	routemap := map[string]*batchedMessage{}
	for _, recipient := range recipients {
		if _, ok := routemap[recipient.Peer]; !ok {
			routemap[recipient.Peer] = &batchedMessage{
				Peer: recipient.Peer,
			}
		}
		routemap[recipient.Peer].Recipients = append(routemap[recipient.Peer].Recipients, recipient.SessionID)
	}
	for peer, batch := range routemap {
		b.mesh.DialAddress("listener", peer, func(conn *grpc.ClientConn) error {
			c := listenerpb.NewClient(conn)
			return c.SendBatchPublish(b.ctx, batch.Recipients, p)
		})
	}
	return nil
}
