package discovery

import (
	"context"

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

func Mesh(id string, logger *zap.Logger, membershipAdapter pb.MembershipAdapter) DiscoveryAdapter {
	return mesh.NewDiscoveryAdapter(id, logger, membershipAdapter)
}
func PB(ctx context.Context, id, host string, logger *zap.Logger) DiscoveryAdapter {
	return pb.NewPBDiscoveryAdapter(ctx, id, host, logger)
}

func Consul(id string, logger *zap.Logger) DiscoveryAdapter {
	return consul.NewConsulDiscoveryAdapter(id, logger)
}
func Static(list []string) DiscoveryAdapter {
	return static.NewStaticDiscoveryAdapter(list)
}
