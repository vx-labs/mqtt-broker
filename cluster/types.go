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
	RegisterService(name, address string) error
	Leave()
	Health() string
}
type DiscoveryLayer interface {
	Members() (peers.SubscriptionSet, error)
	EndpointsByService(name string) ([]*pb.NodeService, error)
	Peers() peers.PeerStore
	DialService(name string) (*grpc.ClientConn, error)
	RegisterService(name, address string) error
	UnregisterService(name string) error
	AddServiceTag(service, key, value string) error
	RemoveServiceTag(name string, tag string) error
	Leave()
	Join(hosts []string) error
	Health() string
	ServiceName() string
}

type DiscoveryAdapter interface {
	Members() (peers.SubscriptionSet, error)
	EndpointsByService(name string) ([]*pb.NodeService, error)
	DialService(name string) (*grpc.ClientConn, error)
	RegisterService(name, address string) error
	UnregisterService(name string) error
	AddServiceTag(service, key, value string) error
	RemoveServiceTag(name string, tag string) error
	Shutdown() error
}

// Mesh represents the mesh state network, being able to broadcast state across the nodes.
type GossipLayer interface {
	AddState(key string, state types.GossipState) (types.Channel, error)
	DiscoverPeers(discovery peers.PeerStore)
	Join(peers []string) error
	Members() []*memberlist.Node
	OnNodeJoin(func(id string, meta []byte))
	OnNodeLeave(func(id string, meta []byte))
	OnNodeUpdate(func(id string, meta []byte))
	Leave()
}
