package sessions

import (
	"context"
	"net"

	"go.uber.org/zap"
)

type server struct {
	id        string
	store     SessionStore
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
