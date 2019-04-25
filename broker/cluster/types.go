package cluster

// State represents a CRDT state store, that will be distributed over the mesh network.
type State interface {
	Merge(inc []byte) error
	MarshalBinary() []byte
}

// Channel allows clients to send messages for a specific state type that will be
// broadcasted in a best-effort manner.
type Channel interface {
	Broadcast(b []byte)
}

// Mesh represents the mesh network, being able to broadcast state across the nodes.
type Mesh interface {
	AddState(key string, state State) (Channel, error)
	Join(hosts []string) error
	MemberRPCAddress(id string) (string, error)
}
