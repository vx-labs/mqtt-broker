package discovery

import (
	"context"
	"net"

	"github.com/vx-labs/mqtt-broker/adapters/discovery/consul"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/nomad"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/static"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type DiscoveryAdapter interface {
	EndpointsByService(name string) ([]*pb.NodeService, error)
	DialService(name string, tags ...string) (*grpc.ClientConn, error)
	ListenTCP(id, name string, port int, advertizedAddress string) (net.Listener, error)
	ListenUDP(id, name string, port int, advertizedAddress string) (net.PacketConn, error)
	Shutdown() error
}

type Service interface {
	DiscoverEndpoints() ([]*pb.NodeService, error)
	Address() string
	ID() string
	Name() string
	BindPort() int
	AdvertisedHost() string
	AdvertisedPort() int
	Dial(tags ...string) (*grpc.ClientConn, error)
	ListenTCP() (net.Listener, error)
	ListenUDP() (net.PacketConn, error)
}

func PB(ctx context.Context, id, host string, logger *zap.Logger) DiscoveryAdapter {
	return pb.NewPBDiscoveryAdapter(ctx, id, host, logger)
}

func Consul(id string, logger *zap.Logger) DiscoveryAdapter {
	return consul.NewConsulDiscoveryAdapter(id, logger)
}
func Nomad(id string, logger *zap.Logger) DiscoveryAdapter {
	return nomad.NewNomadDiscoveryAdapter(id, logger)
}
func Static(list []string) DiscoveryAdapter {
	return static.NewStaticDiscoveryAdapter(list)
}
