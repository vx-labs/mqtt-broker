package router

import (
	"net"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/adapters/identity"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
}
func (b *server) Start(id, name string, mesh discovery.DiscoveryAdapter, catalog identity.Catalog, logger *zap.Logger) error {
	return nil
}
func (m *server) Health() string {
	return "ok"
}
func (m *server) Serve(port int) net.Listener {
	return nil
}
