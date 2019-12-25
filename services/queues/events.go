package queues

import (
	"github.com/pkg/errors"
	"github.com/vx-labs/mqtt-broker/events"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"go.uber.org/zap"
)

func (b *server) consumeStream(messages []*messages.StoredMessage) (int, error) {
	if b.store == nil {
		return 0, errors.New("store not ready")
	}
	for idx := range messages {
		eventSet, err := events.Decode(messages[idx].Payload)
		if err != nil {
			return idx, errors.Wrap(err, "failed to decode message for shard")
		}
		for _, eventPayload := range eventSet {
			switch event := eventPayload.GetEvent().(type) {
			case *events.StateTransition_SessionClosed:
				input := event.SessionClosed
				err = b.clusterDeleteQueue(input.ID)
				if err != nil {
					b.logger.Warn("failed to delete queue", zap.Error(err))
				}
			case *events.StateTransition_SessionLost:
				input := event.SessionLost
				err = b.clusterDeleteQueue(input.ID)
				if err != nil {
					b.logger.Warn("failed to delete queue", zap.Error(err))
				}
			case *events.StateTransition_SessionCreated:
				input := event.SessionCreated
				err := b.clusterCreateQueue(input.ID)
				if err != nil {
					b.logger.Error("failed to create queue", zap.Error(err))
					return idx, err
				}
			}
		}

	}
	return len(messages), nil
}
