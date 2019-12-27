package auth

import (
	"github.com/vx-labs/mqtt-broker/services/auth/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	id         string
	grpcServer *grpc.Server
	logger     *zap.Logger
	store      *store.Static
}

func New(id string, logger *zap.Logger) *server {
	return &server{
		id:     id,
		logger: logger,
		store:  &store.Static{},
	}
}
