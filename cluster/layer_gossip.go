package cluster

import (
	fmt "fmt"
	"log"

	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/layer"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
)

type ServiceConfig struct {
	ID            string
	AdvertiseAddr string
	AdvertisePort int
	BindPort      int
	ServicePort   int
}

func NewGossipServiceLayer(name string, logger *zap.Logger, serviceConfig ServiceConfig, discovery DiscoveryLayer) types.GossipServiceLayer {
	userConfig := config.Config{
		ID:            serviceConfig.ID,
		AdvertiseAddr: serviceConfig.AdvertiseAddr,
		AdvertisePort: serviceConfig.AdvertisePort,
		BindPort:      serviceConfig.BindPort,
	}
	if userConfig.BindPort == 0 {
		log.Fatalf("FATAL: service/%s: user provided 0 as bind port value", name)
	}
	layer := layer.NewGossipLayer(name, logger, userConfig, nil)

	store := discovery.Peers()
	discovery.Peers().On(peers.PeerCreated, func(peers.Peer) {
		layer.DiscoverPeers(store)
	})
	layer.DiscoverPeers(store)
	discovery.RegisterService(name, fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, serviceConfig.ServicePort))
	discovery.RegisterService(fmt.Sprintf("%s_cluster", name), fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, userConfig.AdvertisePort))
	return layer
}
