package discovery

import (
	fmt "fmt"

	"github.com/vx-labs/mqtt-broker/network"

	"google.golang.org/grpc"
)

func (m *discoveryLayer) DialService(name string) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("mesh:///%s", name),
		network.GRPCClientOptions()...,
	)
}
