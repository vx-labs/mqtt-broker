package discovery

import (
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"google.golang.org/grpc"
)

type ServiceCatalog interface {
	Service(name string) Service
	Dial(name string, tags ...string) (*grpc.ClientConn, error)
}

type servicecatalog struct {
	identities identity.Catalog
	adapter    DiscoveryAdapter
}

func (s *servicecatalog) Dial(name string, tags ...string) (*grpc.ClientConn, error) {
	return s.adapter.DialService(name, tags...)
}
func (s *servicecatalog) Service(name string) Service {
	return NewServiceFromIdentity(s.identities.Get(name), s.adapter)
}

func NewServiceCatalog(i identity.Catalog, adapter DiscoveryAdapter) ServiceCatalog {
	return &servicecatalog{
		identities: i,
		adapter:    adapter,
	}
}
