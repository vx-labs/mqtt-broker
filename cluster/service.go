package cluster

import (
	fmt "fmt"
	"log"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
)

type ServiceConfig struct {
	ID            string
	AdvertiseAddr string
	AdvertisePort int
	BindPort      int
	ServicePort   int
}

func NewServiceLayer(name string, serviceConfig ServiceConfig, discovery Mesh) ServiceLayer {
	userConfig := Config{
		ID:            serviceConfig.ID,
		AdvertiseAddr: serviceConfig.AdvertiseAddr,
		AdvertisePort: serviceConfig.AdvertisePort,
		BindPort:      serviceConfig.BindPort,
	}
	if userConfig.BindPort == 0 {
		log.Fatalf("FATAL: service/%s: user provided 0 as bind port value", name)
	}
	layer := NewLayer(name, userConfig, pb.NodeMeta{
		ID: userConfig.ID,
	})

	discovery.Peers().On(PeerCreated, func(Peer) {
		layer.DiscoverPeers(discovery.Peers())
	})
	discovery.RegisterService(name, fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, serviceConfig.ServicePort))
	discovery.RegisterService(fmt.Sprintf("%s_cluster", name), fmt.Sprintf("%s:%d", userConfig.AdvertiseAddr, userConfig.AdvertisePort))
	return layer
}
