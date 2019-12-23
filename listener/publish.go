package listener

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func (b *endpoint) enqueuePublish(ctx context.Context, tenant string, sender string, publish *packet.Publish) error {
	payload := &pb.MessagePublished{
		Tenant:  tenant,
		Publish: publish,
	}
	data, err := proto.Marshal(payload)
	if err != nil {
		return err
	}
	err = b.messages.Put(ctx, "messages", sender, data)
	if err != nil {
		b.logger.Error("failed to enqueue message in message store", zap.Error(err))
		return err
	}
	return nil
}

func (local *endpoint) PublishHandler(ctx context.Context, session *localSession, p *packet.Publish) error {
	decodedToken, err := DecodeSessionToken(SigningKey(), session.token)
	if err != nil {
		session.logger.Error("failed to validate session token", zap.Error(err))
		return err
	}
	err = local.enqueuePublish(ctx, decodedToken.SessionTenant, decodedToken.SessionID, p)
	if err != nil {
		session.logger.Error("failed to publish packet", zap.Error(err))
		return err
	}
	switch p.Header.Qos {
	case 1:
		return session.encoder.PubAck(&packet.PubAck{Header: &packet.Header{}, MessageId: p.MessageId})
	default:
		return nil
	}
}
