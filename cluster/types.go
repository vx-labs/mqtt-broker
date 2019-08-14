package cluster

import (
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"google.golang.org/grpc"
)

// Mesh represents  the mesh discovery network.
type Mesh interface {
	Join(hosts []string) error
	Peers() peers.PeerStore
	DialService(name string) (*grpc.ClientConn, error)
	DialAddress(service, id string, f func(*grpc.ClientConn) error) error
	RegisterService(name, address string) error
	Leave()
	Health() string
}
type DiscoveryLayer interface {
	Peers() peers.PeerStore
	DialService(name string) (*grpc.ClientConn, error)
	DialAddress(service, id string, f func(*grpc.ClientConn) error) error
	RegisterService(name, address string) error
	Leave()
	Join(hosts []string) error
	Health() string
}

// Mesh represents the mesh state network, being able to broadcast state across the nodes.
type GossipLayer interface {
	AddState(key string, state types.GossipState) (types.Channel, error)
	DiscoverPeers(discovery peers.PeerStore)
	Join(peers []string) error
	Members() []*memberlist.Node
	OnNodeJoin(func(id string, meta pb.NodeMeta))
	OnNodeLeave(func(id string, meta pb.NodeMeta))
	Leave()
}
