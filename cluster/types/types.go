package types

import "github.com/vx-labs/mqtt-broker/cluster/pb"

// GossipState represents a CRDT state store, that will be distributed over the mesh network.
type GossipState interface {
	Merge(inc []byte, full bool) error
	MarshalBinary() []byte
}

// Channel allows clients to send messages for a specific state type that will be
// broadcasted in a best-effort manner.
type Channel interface {
	Broadcast(b []byte)
}

type GossipServiceLayer interface {
	AddState(key string, state GossipState) (Channel, error)
	OnNodeLeave(f func(id string, meta pb.NodeMeta))
}
