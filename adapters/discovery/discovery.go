package discovery

import (
	"github.com/vx-labs/mqtt-broker/adapters/discovery/mesh"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type DiscoveryAdapter interface {
	Members() ([]*pb.Peer, error)
	EndpointsByService(name string) ([]*pb.NodeService, error)
	DialService(name string) (*grpc.ClientConn, error)
	RegisterService(name, address string) error
	UnregisterService(name string) error
	AddServiceTag(service, key, value string) error
	RemoveServiceTag(name string, tag string) error
	Shutdown() error
}

func Mesh(logger *zap.Logger, userConfig config.Config) DiscoveryAdapter {
	return mesh.NewDiscoveryAdapter(logger, userConfig)
}
