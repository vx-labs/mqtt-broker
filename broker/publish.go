package broker

import (
	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func (b *Broker) enqueuePublish(tenant string, publish *packet.Publish) {
	// payload := &publishQueue.Message{
	// 	Tenant:  tenant,
	// 	Publish: publish,
	// }
	data, err := proto.Marshal(publish)
	if err == nil {
		err := b.Messages.Put(b.ctx, "messages", tenant, data)
		if err != nil {
			b.logger.Error("failed to enqueue message in message store", zap.Error(err))
		}
	}
	//b.publishQueue.Enqueue(payload)
}
