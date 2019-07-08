package broker

import (
	"net"

	"github.com/vx-labs/mqtt-broker/cluster"
)

func (b *Broker) Serve(port int) net.Listener {
	return Serve(port, b)
}
func (b *Broker) Shutdown() {
	b.Stop()
}
func (b *Broker) JoinServiceLayer(layer cluster.ServiceLayer) {
	b.Start(layer)
}
