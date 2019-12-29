package cluster

import (
	fmt "fmt"
	"log"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/layer"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
)

func NewRaftServiceLayer(name string, logger *zap.Logger, serviceConfig ServiceConfig, rpcConfig ServiceConfig, discovery discovery.DiscoveryAdapter) types.RaftServiceLayer {
	userConfig := config.Config{
		ID:            serviceConfig.ID,
		AdvertiseAddr: serviceConfig.AdvertiseAddr,
		AdvertisePort: serviceConfig.AdvertisePort,
		BindPort:      serviceConfig.BindPort,
	}
	if userConfig.BindPort == 0 {
		log.Fatalf("FATAL: service/%s: user provided 0 as bind port value", name)
	}
	l, err := layer.NewRaftLayer(logger, userConfig, discovery)
	if err != nil {
		logger.Error("raft starting failed", zap.Error(err))
	}

	discovery.RegisterService(name, fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, serviceConfig.ServicePort))
	discovery.RegisterService(fmt.Sprintf("%s_rpc", name), fmt.Sprintf("%s:%d", rpcConfig.AdvertiseAddr, rpcConfig.AdvertisePort))
	ServeRPC(rpcConfig.BindPort, l)
	return l
}
