package cluster

import (
	fmt "fmt"
	"log"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
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

func NewServiceLayer(name string, logger *zap.Logger, serviceConfig ServiceConfig, discovery Mesh) types.ServiceLayer {
	userConfig := Config{
		ID:            serviceConfig.ID,
		AdvertiseAddr: serviceConfig.AdvertiseAddr,
		AdvertisePort: serviceConfig.AdvertisePort,
		BindPort:      serviceConfig.BindPort,
	}
	if userConfig.BindPort == 0 {
		log.Fatalf("FATAL: service/%s: user provided 0 as bind port value", name)
	}
	layer := NewLayer(name, logger, userConfig, pb.NodeMeta{
		ID: userConfig.ID,
	})

	store := discovery.Peers()
	discovery.Peers().On(peers.PeerCreated, func(peers.Peer) {
		layer.DiscoverPeers(store)
	})
	layer.DiscoverPeers(store)
	discovery.RegisterService(name, fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, serviceConfig.ServicePort))
	discovery.RegisterService(fmt.Sprintf("%s_cluster", name), fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, userConfig.AdvertisePort))
	return layer
}
