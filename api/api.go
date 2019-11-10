package api

import (
	"context"
	"net"

	"go.uber.org/zap"

	brokerClient "github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/cluster"
	sessionsClient "github.com/vx-labs/mqtt-broker/sessions/pb"
	subscriptionsClient "github.com/vx-labs/mqtt-broker/subscriptions/pb"
)

type api struct {
	id                  string
	ctx                 context.Context
	config              Config
	listeners           []net.Listener
	mesh                cluster.Mesh
	brokerClient        *brokerClient.Client
	sessionsClient      *sessionsClient.Client
	subscriptionsClient *subscriptionsClient.Client
	logger              *zap.Logger
}
type Config struct {
	TlsCommonName string
	TlsPort       int
	TcpPort       int
	logger        *zap.Logger
}

func New(id string, logger *zap.Logger, mesh cluster.Mesh, config Config) *api {
	brokerConn, err := mesh.DialService("broker")
	if err != nil {
		panic(err)
	}
	sessionsConn, err := mesh.DialService("sessions?raft_status=leader")
	if err != nil {
		panic(err)
	}
	subscriptionsConn, err := mesh.DialService("subscriptions?raft_status=leader")
	if err != nil {
		panic(err)
	}
	return &api{
		id:                  id,
		ctx:                 context.Background(),
		mesh:                mesh,
		config:              config,
		brokerClient:        brokerClient.NewClient(brokerConn),
		sessionsClient:      sessionsClient.NewClient(sessionsConn),
		subscriptionsClient: subscriptionsClient.NewClient(subscriptionsConn),
		logger:              logger,
	}
}
