package cluster

import "github.com/vx-labs/mqtt-broker/cluster/peers"

type meshTransport struct {
	id    string
	peers peers.PeerStore
	mesh  Mesh
}
