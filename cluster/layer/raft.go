package layer

import (
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/layer/consistency"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type DiscoveryProvider interface {
	RegisterService(string, string) error
	UnregisterService(string) error
	AddServiceTag(service, key, value string) error
	RemoveServiceTag(name string, tag string) error
	DialService(id string) (*grpc.ClientConn, error)
	Peers() peers.PeerStore
}

func NewRaftLayer(logger *zap.Logger, userConfig config.Config, discovery DiscoveryProvider) (types.RaftServiceLayer, error) {
	return consistency.New(logger, userConfig, discovery)
}
