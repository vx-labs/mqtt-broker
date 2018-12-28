package peers

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/peers/ --go_out=plugins=grpc:. types.proto

type peerFilter func(*Peer) bool

func (set PeerList) Filter(filters ...peerFilter) PeerList {
	copy := make([]*Peer, 0, len(set.Peers))
	for _, peer := range set.Peers {
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
	return PeerList{
		Peers: copy,
	}
}
func (set PeerList) Apply(f func(s *Peer)) {
	for _, peer := range set.Peers {
		f(peer)
	}
}
func (set PeerList) ApplyIdx(f func(idx int, s *Peer)) {
	for idx, peer := range set.Peers {
		f(idx, peer)
	}
}

func (set PeerList) ApplyE(f func(s *Peer) error) error {
	for _, peer := range set.Peers {
		if err := f(peer); err != nil {
			return err
		}
	}
	return nil
}

func (s *Peer) IsAdded() bool {
	return s.LastAdded > 0 && s.LastAdded > s.LastDeleted
}
func (s *Peer) IsRemoved() bool {
	return s.LastDeleted > 0 && s.LastAdded < s.LastDeleted
}

func IsOutdated(s *Peer, remote *Peer) (outdated bool) {
	if s.LastAdded < remote.LastAdded {
		outdated = true
	}
	if s.LastDeleted < remote.LastDeleted {
		outdated = true
	}
	return outdated
}
