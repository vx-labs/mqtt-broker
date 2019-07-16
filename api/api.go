package api

import (
	"crypto/tls"
	"net"

	brokerClient "github.com/vx-labs/mqtt-broker/broker/pb"
	"github.com/vx-labs/mqtt-broker/cluster"
)

type api struct {
	config       Config
	listeners    []net.Listener
	mesh         cluster.Mesh
	brokerClient *brokerClient.Client
}
type Config struct {
	TlsConfig *tls.Config
	TlsPort   int
	TcpPort   int
}

func New(id string, mesh cluster.Mesh, config Config) *api {
	conn, err := mesh.DialService("broker")
	if err != nil {
		panic(err)
	}
	return &api{
		mesh:         mesh,
		config:       config,
		brokerClient: brokerClient.NewClient(conn),
	}
}
