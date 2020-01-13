package discovery

import (
	"context"
	"net"

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
	RegisterTCPService(id, name, address string) error
	RegisterUDPService(id, name, address string) error
	RegisterGRPCService(id, name, address string) error
	UnregisterService(id string) error
	AddServiceTag(id, key, value string) error
	RemoveServiceTag(id string, tag string) error
	Shutdown() error
}

type Service interface {
	DiscoverEndpoints() ([]*pb.NodeService, error)
	RegisterTCP() error
	RegisterGRPC() error
	Unregister() error
	Address() string
	ID() string
	Name() string
	BindPort() int
	AdvertisedHost() string
	AdvertisedPort() int
	AddTag(key, value string) error
	RemoveTag(tag string) error
	Dial(tags ...string) (*grpc.ClientConn, error)
	ListenTCP() (net.Listener, error)
	ListenUDP() (net.PacketConn, error)
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
