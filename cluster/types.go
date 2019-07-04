package cluster

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/broker/cluster/ --go_out=plugins=grpc:. cluster.proto

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
	Peers() PeerStore
}

type peerFilter func(Peer) bool
type SubscriptionSet []Peer

func (set SubscriptionSet) Filter(filters ...peerFilter) SubscriptionSet {
	copy := make(SubscriptionSet, 0, len(set))
	for _, peer := range set {
		accepted := true
		for _, f := range filters {
			if !f(peer) {
				accepted = false
				break
			}
		}
		if accepted {
			copy = append(copy, peer)
		}
	}
	return copy
}
func (set SubscriptionSet) Apply(f func(s Peer)) {
	for _, peer := range set {
		f(peer)
	}
}
func (set SubscriptionSet) ApplyIdx(f func(idx int, s Peer)) {
	for idx, peer := range set {
		f(idx, peer)
	}
}

func (set SubscriptionSet) ApplyE(f func(s Peer) error) error {
	for _, peer := range set {
		if err := f(peer); err != nil {
			return err
		}
	}
	return nil
}
