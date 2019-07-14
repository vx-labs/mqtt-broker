package cluster

import (
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"google.golang.org/grpc"
)

// Mesh represents  the mesh discovery network.
type Mesh interface {
	Join(hosts []string)
	Peers() peers.PeerStore
	DialService(name string) (*grpc.ClientConn, error)
	DialAddress(service, id string, f func(*grpc.ClientConn) error) error
	RegisterService(name, address string) error
	Leave()
	Health() string
}

// Mesh represents the mesh state network, being able to broadcast state across the nodes.
type Layer interface {
	AddState(key string, state types.State) (types.Channel, error)
	DiscoverPeers(discovery peers.PeerStore)
	Join(peers []string)
	Members() []*memberlist.Node
	Leave()
}
