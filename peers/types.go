package peers

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/peers/ --go_out=plugins=grpc:. types.proto

type peerFilter func(*Metadata) bool

func (set PeerMetadataList) Filter(filters ...peerFilter) PeerMetadataList {
	copy := make([]*Metadata, 0, len(set.Peers))
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
	return PeerMetadataList{
		Peers: copy,
	}
}
func (set PeerMetadataList) Apply(f func(s *Metadata)) {
	for _, peer := range set.Peers {
		f(peer)
	}
}
func (set PeerMetadataList) ApplyIdx(f func(idx int, s *Metadata)) {
	for idx, peer := range set.Peers {
		f(idx, peer)
	}
}

func (set PeerMetadataList) ApplyE(f func(s *Metadata) error) error {
	for _, peer := range set.Peers {
		if err := f(peer); err != nil {
			return err
		}
	}
	return nil
}

func (s *Metadata) IsAdded() bool {
	return s.LastAdded > 0 && s.LastAdded > s.LastDeleted
}
func (s *Metadata) IsRemoved() bool {
	return s.LastDeleted > 0 && s.LastAdded < s.LastDeleted
}

func IsOutdated(s *Metadata, remote *Metadata) (outdated bool) {
	if s.LastAdded < remote.LastAdded {
		outdated = true
	}
	if s.LastDeleted < remote.LastDeleted {
		outdated = true
	}
	return outdated
}
