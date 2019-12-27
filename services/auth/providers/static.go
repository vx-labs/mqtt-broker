package providers

import "github.com/vx-labs/mqtt-broker/services/auth/pb"

type static struct {
	keys map[string]string
}

func (s *static) Authenticate(tenant, id string, p *pb.ProtocolContext, t *pb.TransportContext) bool {
	key, ok := s.keys[tenant]
	if !ok {
		return false
	}
	return p.Password == key
}
