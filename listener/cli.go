package listener

import (
	"net"

	"github.com/vx-labs/mqtt-broker/cluster"
)

func (b *endpoint) Serve(port int) net.Listener {
	return Serve(b, port)
}
func (b *endpoint) Shutdown() {
	b.Close()
}
func (b *endpoint) JoinServiceLayer(layer cluster.ServiceLayer) {
}
