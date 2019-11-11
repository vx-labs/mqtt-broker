package cluster

import (
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/discovery"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

func NewDiscoveryLayer(logger *zap.Logger, userConfig config.Config) DiscoveryLayer {
	layer := discovery.NewDiscoveryLayer(logger, userConfig)
	resolver.Register(newResolver(layer.Peers(), logger))
	return layer
}
