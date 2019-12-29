package listener

import (
	"net"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"go.uber.org/zap"
)

func (b *endpoint) Serve(port int) net.Listener {
	return nil
}
func (b *endpoint) Shutdown() {
	b.Close()
}
func (b *endpoint) Start(id, name string, mesh discovery.DiscoveryAdapter, catalog identity.Catalog, logger *zap.Logger) error {
	return nil
}
func (m *endpoint) Health() string {
	return "ok"
}
