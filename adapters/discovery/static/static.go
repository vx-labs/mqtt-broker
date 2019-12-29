package static

import (
	"errors"
	"io"

	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"google.golang.org/grpc"
)

type StaticDiscoveryAdapter struct {
	list []string
}

func NewStaticDiscoveryAdapter(list []string) *StaticDiscoveryAdapter {
	return &StaticDiscoveryAdapter{
		list: list,
	}
}

func (c *StaticDiscoveryAdapter) EndpointsByService(name string) ([]*pb.NodeService, error) {
	if len(c.list) == 0 {
		return nil, io.EOF
	}
	out := make([]*pb.NodeService, len(c.list))
	for idx, address := range c.list {
		out[idx] = &pb.NodeService{
			ID:             name,
			Peer:           address,
			NetworkAddress: address,
			Tags:           nil,
		}
	}
	return out, nil
}
func (c *StaticDiscoveryAdapter) RegisterService(name, address string) error {
	return errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) UnregisterService(name string) error {
	return errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) AddServiceTag(service, key, value string) error {
	return errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) RemoveServiceTag(service, key string) error {
	return errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) DialService(name string, tags ...string) (*grpc.ClientConn, error) {
	return nil, errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) Shutdown() error {
	return nil
}
func (c *StaticDiscoveryAdapter) Members() ([]*pb.Peer, error) {
	return nil, errors.New("Unsupported yet")
}
