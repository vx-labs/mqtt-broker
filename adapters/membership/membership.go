package membership

import (
	discovery "github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/adapters/membership/mesh"
	"go.uber.org/zap"
)

// MembershipAdapter discovers a group of nodes sharing a common attribute, and allow to observe changes on this group.
// It is responsible to handle new nodes discovery, failed nodes evinction and node's metadata updates propagations.
type MembershipAdapter interface {
	Shutdown() error
	UpdateMetadata([]byte)
	OnNodeJoin(f func(id string, meta []byte))
	OnNodeUpdate(f func(id string, meta []byte))
	OnNodeLeave(f func(id string, meta []byte))
	MemberCount() int
}

type DiscoveryAdapter interface {
	DiscoverEndpoints() ([]*discovery.NodeService, error)
	Address() string
	AdvertisedHost() string
	AdvertisedPort() int
	BindPort() int
}

func Mesh(id string, logger *zap.Logger, service DiscoveryAdapter) *mesh.GossipMembershipAdapter {
	return mesh.NewMesh(id, logger, service)
}
