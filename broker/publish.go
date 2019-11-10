package broker

import (
	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func (b *Broker) enqueuePublish(tenant string, publish *packet.Publish) error {
	payload := &pb.MessagePublished{
		Tenant:  tenant,
		Publish: publish,
	}
	data, err := proto.Marshal(payload)
	if err != nil {
		return err
	}
	err = b.Messages.Put(b.ctx, "messages", tenant, data)
	if err != nil {
		b.logger.Error("failed to enqueue message in message store", zap.Error(err))
		return err
	}
	return nil
}
