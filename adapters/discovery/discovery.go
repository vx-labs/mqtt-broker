package discovery

import (
	"github.com/vx-labs/mqtt-broker/adapters/discovery/consul"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/mesh"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/static"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type DiscoveryAdapter interface {
	Members() ([]*pb.Peer, error)
	EndpointsByService(name string) ([]*pb.NodeService, error)
	DialService(name string, tags ...string) (*grpc.ClientConn, error)
	RegisterService(name, address string) error
	UnregisterService(name string) error
	AddServiceTag(service, key, value string) error
	RemoveServiceTag(name string, tag string) error
	Shutdown() error
}

type Service interface {
	DiscoverEndpoints() ([]*pb.NodeService, error)
	Register() error
	Unregister() error
	Address() string
	Name() string
	BindPort() int
	AdvertisedHost() string
	AdvertisedPort() int
	AddTag(key, value string) error
	RemoveTag(tag string) error
	Dial(tags ...string) (*grpc.ClientConn, error)
}

func Mesh(id string, logger *zap.Logger, membershipAdapter pb.MembershipAdapter) *mesh.MeshDiscoveryAdapter {
	return mesh.NewDiscoveryAdapter(id, logger, membershipAdapter)
}

func Consul(id string, logger *zap.Logger) *consul.ConsulDiscoveryAdapter {
	return consul.NewConsulDiscoveryAdapter(id, logger)
}
func Static(list []string) *static.StaticDiscoveryAdapter {
	return static.NewStaticDiscoveryAdapter(list)
}
