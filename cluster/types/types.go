package types

import "github.com/vx-labs/mqtt-broker/cluster/pb"

// State represents a CRDT state store, that will be distributed over the mesh network.
type State interface {
	Merge(inc []byte, full bool) error
	MarshalBinary() []byte
}

// Channel allows clients to send messages for a specific state type that will be
// broadcasted in a best-effort manner.
type Channel interface {
	Broadcast(b []byte)
}

type ServiceLayer interface {
	AddState(key string, state State) (Channel, error)
	OnNodeLeave(f func(id string, meta pb.NodeMeta))
}
