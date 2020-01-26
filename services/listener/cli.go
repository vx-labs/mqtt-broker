package listener

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/events"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/stream"
	"go.uber.org/zap"
)

func (b *endpoint) Serve(port int) net.Listener {
	return b.listener
}
func (b *endpoint) Shutdown() {
	b.Close()
	b.streamClient.Shutdown()
}
func (b *endpoint) Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error {
	listener, err := catalog.Service(name).ListenTCP()
	if err != nil {
		return err
	}
	b.listener = listener
	ctx := context.Background()
	b.streamClient = stream.NewClient(b.kv, b.messages, logger)
	b.streamClient.ConsumeStream(ctx, "events", b.consumeStream,
		stream.WithConsumerID(b.id),
		stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_NOW),
	)

	return nil
}
func (m *endpoint) Health() string {
	return "ok"
}

func (b *endpoint) consumeStream(messages []*messages.StoredMessage) (int, error) {
	for idx := range messages {
		eventSet, err := events.Decode(messages[idx].Payload)
		if err != nil {
			return idx, errors.Wrap(err, "failed to decode message for shard")
		}
		for _, eventPayload := range eventSet {
			switch event := eventPayload.GetEvent().(type) {
			case *events.StateTransition_SessionClosed:
				input := event.SessionClosed
				b.mutex.Lock()
				res := b.sessions.Delete(&localSession{id: input.ID})
				b.mutex.Unlock()
				if res != nil {
					res.(*localSession).cancel()
				}
			case *events.StateTransition_SessionLost:
				input := event.SessionLost
				b.mutex.Lock()
				res := b.sessions.Delete(&localSession{id: input.ID})
				b.mutex.Unlock()
				if res != nil {
					res.(*localSession).cancel()
				}
			}
		}
	}
	return len(messages), nil
}
