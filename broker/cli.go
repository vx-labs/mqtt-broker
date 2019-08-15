package broker

import (
	"net"

	"github.com/vx-labs/mqtt-broker/cluster"
	"go.uber.org/zap"
)

func (b *Broker) Serve(port int) net.Listener {
	return Serve(port, b)
}
func (b *Broker) Shutdown() {
	b.Stop()
}
func (b *Broker) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	l := cluster.NewGossipServiceLayer(name, logger, config, mesh)
	b.Start(l)
}

func (b *Broker) Health() string {
	return "ok"
}
