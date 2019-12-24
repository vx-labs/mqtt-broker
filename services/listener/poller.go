package listener

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func (local *endpoint) startPoller(ctx context.Context, session *localSession, poller chan error) {
	defer close(poller)
	session.logger.Debug("starting queue poller")
	count := 0
	err := backoff.Retry(func() error {
		err := local.queues.StreamMessages(ctx, session.id, func(ackOffset uint64, message *packet.Publish) error {
			if message == nil {
				session.logger.Error("received empty message from queue")
				return nil
			}
			if message.Header == nil {
				message.Header = &packet.Header{Qos: 0}
			}
			switch message.Header.Qos {
			case 0:
				message.MessageId = 1
				session.encoder.Publish(message)
				err := local.queues.AckMessage(ctx, session.id, ackOffset)
				if err != nil {
					select {
					case <-ctx.Done():
						return nil
					default:
					}
					session.logger.Error("failed to ack message on queues service", zap.Error(err))
				}
			case 1:
				session.inflights.Put(ctx, message, func() {
					err := local.queues.AckMessage(ctx, session.id, ackOffset)
					if err != nil {
						select {
						case <-ctx.Done():
							return
						default:
						}
						session.logger.Error("failed to ack message on queues service", zap.Error(err))
					}
				})
			default:
			}
			return nil
		})
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if count >= 5 {
			return backoff.Permanent(err)
		}
		count++
		return err
	}, backoff.NewExponentialBackOff())
	if err != nil {
		session.logger.Error("failed to poll queues for messages", zap.Error(err))
	}
	poller <- err
}
