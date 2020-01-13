package listener

import (
	"context"
	"time"

	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func (local *endpoint) SubscribeHandler(ctx context.Context, session *localSession, p *packet.Subscribe) error {
	waitTicker := time.NewTicker(1 * time.Second)
	defer waitTicker.Stop()
	retries := 3
	var err error
	for retries > 0 {
		retries--
		_, err = local.queues.GetQueueStatistics(ctx, session.id)
		if err == nil {
			break
		}
		<-waitTicker.C
	}
	if err != nil {
		session.logger.Error("timed out waiting for queue to be created", zap.Error(err))
		return err
	}
	session.poller.Do(func() {
		go local.startPoller(ctx, session, session.pollerCh)
	})
	decodedToken, err := DecodeSessionToken(SigningKey(), session.token)
	if err != nil {
		session.logger.Error("failed to validate session token", zap.Error(err))
		return err
	}
	eventSet := []*events.StateTransition{}
	for idx := range p.Topic {
		pattern := p.Topic[idx]
		qos := p.Qos[idx]
		eventSet = append(eventSet, &events.StateTransition{
			Event: &events.StateTransition_SessionSubscribed{
				SessionSubscribed: &events.SessionSubscribed{
					Pattern:   pattern,
					Qos:       qos,
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
	return session.encoder.SubAck(&packet.SubAck{
		Header:    &packet.Header{},
		MessageId: p.MessageId,
		Qos:       p.Qos,
	})

}
