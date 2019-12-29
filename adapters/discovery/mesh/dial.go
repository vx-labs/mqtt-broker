package mesh

import (
	fmt "fmt"
	"strings"

	"github.com/vx-labs/mqtt-broker/network"

	"google.golang.org/grpc"
)

func (m *MeshDiscoveryAdapter) Shutdown() error {
	m.Leave()
	return nil
}
func (m *MeshDiscoveryAdapter) DialService(name string, tags ...string) (*grpc.ClientConn, error) {
	key := fmt.Sprintf("mesh:///%s", name)
	if len(tags) > 0 {
		key = fmt.Sprintf("%s?%s", key, strings.Join(tags, "&"))
	}
	return grpc.Dial(key,
		network.GRPCClientOptions()...,
	)
}
