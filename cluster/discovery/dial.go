package discovery

import (
	"fmt"

	"github.com/vx-labs/mqtt-broker/network"

	"google.golang.org/grpc"
)

func (m *discoveryLayer) DialService(name string) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("mesh:///%s", name),
		network.GRPCClientOptions()...,
	)
}
func (m *discoveryLayer) DialAddress(service, id string, f func(*grpc.ClientConn) error) error {
	return m.rpcCaller.Call(fmt.Sprintf("%s+%s", service, id), f)
}
