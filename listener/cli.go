package listener

import (
	"net"

	"github.com/vx-labs/mqtt-broker/cluster"
	"go.uber.org/zap"
)

func (b *endpoint) Serve(port int) net.Listener {
	return nil
}
func (b *endpoint) Shutdown() {
	b.Close()
}
func (b *endpoint) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
}
func (m *endpoint) Health() string {
	return "ok"
}
