package cp

import (
	"github.com/vx-labs/mqtt-broker/adapters/cp/pb"
	"github.com/vx-labs/mqtt-broker/adapters/cp/raft"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"go.uber.org/zap"
)

type Synchronizer interface {
	Shutdown() error
	Health() (string, string)
	ApplyEvent(event []byte) error
	IsLeader() bool
}

func RaftSynchronizer(id string, userService, raftService discovery.Service, rpc discovery.Service, state pb.CPState, logger *zap.Logger) Synchronizer {
	return raft.NewRaftSynchronizer(id, userService, raftService, rpc, state, logger)
}
