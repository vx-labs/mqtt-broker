package api

import (
	"context"
	"net"

	"go.uber.org/zap"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	brokerClient "github.com/vx-labs/mqtt-broker/services/broker/pb"
	sessionsClient "github.com/vx-labs/mqtt-broker/services/sessions/pb"
	subscriptionsClient "github.com/vx-labs/mqtt-broker/services/subscriptions/pb"
)

type api struct {
	id                  string
	ctx                 context.Context
	config              Config
	listeners           []net.Listener
	mesh                discovery.DiscoveryAdapter
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

func New(id string, logger *zap.Logger, mesh discovery.DiscoveryAdapter, config Config) *api {
	brokerConn, err := mesh.DialService("broker")
	if err != nil {
		panic(err)
	}
	sessionsConn, err := mesh.DialService("sessions")
	if err != nil {
		panic(err)
	}
	subscriptionsConn, err := mesh.DialService("subscriptions")
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
