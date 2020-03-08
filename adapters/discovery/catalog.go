package discovery

import (
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"google.golang.org/grpc"
)

type ServiceCatalog interface {
	Service(name, tag string) Service
	Dial(name string, tag string) (*grpc.ClientConn, error)
}

type servicecatalog struct {
	identities identity.Catalog
	adapter    DiscoveryAdapter
}

func (s *servicecatalog) Dial(name string, tag string) (*grpc.ClientConn, error) {
	return s.adapter.DialService(name, tag)
}
func (s *servicecatalog) Service(name, tag string) Service {
	return NewServiceFromIdentity(s.identities.Get(name, tag), s.adapter)
}

func NewServiceCatalog(i identity.Catalog, adapter DiscoveryAdapter) ServiceCatalog {
	return &servicecatalog{
		identities: i,
		adapter:    adapter,
	}
}
