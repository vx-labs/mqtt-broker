package sessions

import (
	"context"
	"net"

	"github.com/vx-labs/mqtt-broker/cluster/types"
	"go.uber.org/zap"
)

type server struct {
	id        string
	store     SessionStore
	state     types.RaftServiceLayer
	ctx       context.Context
	listeners []net.Listener
	logger    *zap.Logger
}

func New(id string, logger *zap.Logger) *server {

	return &server{
		id:     id,
		ctx:    context.Background(),
		logger: logger,
	}
}
