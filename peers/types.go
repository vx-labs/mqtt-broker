package peers

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/peers/ --go_out=plugins=grpc:. types.proto

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
