package static

import (
	"errors"
	"io"
	"net"

	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/network"
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

func (c *StaticDiscoveryAdapter) EndpointsByService(name, tag string) ([]*pb.NodeService, error) {
	if len(c.list) == 0 {
		return nil, io.EOF
	}
	out := make([]*pb.NodeService, len(c.list))
	for idx, address := range c.list {
		out[idx] = &pb.NodeService{
			ID:             name,
			Peer:           address,
			NetworkAddress: address,
			Tag:            tag,
		}
	}
	return out, nil
}
func (c *StaticDiscoveryAdapter) RegisterGRPCService(id, name, address string) error {
	return errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) RegisterTCPService(id, name, address string) error {
	return errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) RegisterUDPService(id, name, address string) error {
	return errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) UnregisterService(name string) error {
	return errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) DialService(name string, tag string) (*grpc.ClientConn, error) {
	return grpc.Dial(c.list[0],
		network.GRPCClientOptions()...,
	)
}
func (c *StaticDiscoveryAdapter) ListenTCP(id, name string, port int, advertizedAddress string) (net.Listener, error) {
	return nil, errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) ListenUDP(id, name string, port int, advertizedAddress string) (net.PacketConn, error) {
	return nil, errors.New("Unsupported yet")
}
func (c *StaticDiscoveryAdapter) Shutdown() error {
	return nil
}
func (c *StaticDiscoveryAdapter) Members() ([]*pb.Peer, error) {
	return nil, errors.New("Unsupported yet")
}
