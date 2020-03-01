package ap

import (
	"github.com/vx-labs/mqtt-broker/adapters/ap/gossip"
	"github.com/vx-labs/mqtt-broker/adapters/ap/pb"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"go.uber.org/zap"
)

type Distributer interface {
	Shutdown() error
	Health() (string, string)
}

func GossipDistributer(id string, service discovery.Service, state pb.APState, logger *zap.Logger) Distributer {
	return gossip.NewGossipDistributer(id, service, state, logger)
}
