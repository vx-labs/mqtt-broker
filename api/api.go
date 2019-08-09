package api

import (
	"context"
	"net"

	"go.uber.org/zap"

	brokerClient "github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/cluster"
)

type api struct {
	id           string
	ctx          context.Context
	config       Config
	listeners    []net.Listener
	mesh         cluster.Mesh
	brokerClient *brokerClient.Client
	logger       *zap.Logger
}
type Config struct {
	TlsCommonName string
	TlsPort       int
	TcpPort       int
	logger        *zap.Logger
}

func New(id string, logger *zap.Logger, mesh cluster.Mesh, config Config) *api {
	conn, err := mesh.DialService("broker")
	if err != nil {
		panic(err)
	}
	return &api{
		id:           id,
		ctx:          context.Background(),
		mesh:         mesh,
		config:       config,
		brokerClient: brokerClient.NewClient(conn),
		logger:       logger,
	}
}
