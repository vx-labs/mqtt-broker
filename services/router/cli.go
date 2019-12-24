package router

import (
	"net"

	"github.com/vx-labs/mqtt-broker/cluster"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
}
func (m *server) Health() string {
	return "ok"
}
func (m *server) Serve(port int) net.Listener {
	return nil
}
