package router

import (
	"net"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"github.com/vx-labs/mqtt-broker/stream"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	b.stream.Shutdown()
}
func (b *server) Start(id, name string, mesh discovery.DiscoveryAdapter, catalog identity.Catalog, logger *zap.Logger) error {
	b.stream.ConsumeStream(b.ctx, "messages", b.v2ConsumePayload,
		stream.WithConsumerID(b.id),
		stream.WithConsumerGroupID("router"),
		stream.WithMaxBatchSize(200),
		stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
	)
	return nil
}
func (m *server) Health() string {
	return "ok"
}
func (m *server) Serve(port int) net.Listener {
	return nil
}
