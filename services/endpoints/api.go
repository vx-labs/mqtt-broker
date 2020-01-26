package endpoints

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
)

type api struct {
	id         string
	ctx        context.Context
	grpcServer *grpc.Server
	mesh       discovery.DiscoveryAdapter
	listener   net.Listener
	logger     *zap.Logger
}

func New(id string, logger *zap.Logger, mesh discovery.DiscoveryAdapter) *api {
	return &api{
		id:     id,
		ctx:    context.Background(),
		mesh:   mesh,
		logger: logger,
	}
}
