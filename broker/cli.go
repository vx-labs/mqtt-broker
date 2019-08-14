package broker

import (
	"net"

	"github.com/vx-labs/mqtt-broker/cluster/types"
)

func (b *Broker) Serve(port int) net.Listener {
	return Serve(port, b)
}
func (b *Broker) Shutdown() {
	b.Stop()
}
func (b *Broker) JoinServiceLayer(layer types.GossipServiceLayer) {
	b.Start(layer)
}
func (b *Broker) Health() string {
	return "ok"
}
