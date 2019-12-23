package listener

import (
	"context"

	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func (local *endpoint) UnsubscribeHandler(ctx context.Context, session *localSession, p *packet.Unsubscribe) error {
	decodedToken, err := DecodeSessionToken(SigningKey(), session.token)
	if err != nil {
		session.logger.Error("failed to validate session token", zap.Error(err))
		return err
	}
	eventSet := []*events.StateTransition{}
	for idx := range p.Topic {
		pattern := p.Topic[idx]
		eventSet = append(eventSet, &events.StateTransition{
			Event: &events.StateTransition_SessionUnsubscribed{
				SessionUnsubscribed: &events.SessionUnsubscribed{
					Pattern:   pattern,
					SessionID: decodedToken.SessionID,
					Tenant:    decodedToken.SessionTenant,
				},
			},
		})
	}
	payload, err := events.Encode(eventSet...)
	if err != nil {
		return err
	}
	err = local.messages.Put(ctx, "events", decodedToken.SessionID, payload)
	if err != nil {
		local.logger.Error("failed to enqueue event in message store", zap.Error(err))
		return err
	}
	return session.encoder.UnsubAck(&packet.UnsubAck{
		Header:    &packet.Header{},
		MessageId: p.MessageId,
	})

}
