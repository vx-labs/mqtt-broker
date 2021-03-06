package discovery

import (
	"fmt"
	"net"
	"strconv"

	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"google.golang.org/grpc"
)

type service struct {
	health   string
	adapter  DiscoveryAdapter
	id       string
	name     string
	tag      string
	address  string
	bindPort int
}

func (s *service) ListenTCP() (net.Listener, error) {
	return s.adapter.ListenTCP(s.ID(), s.Name(), s.BindPort(), s.Address())
}
func (s *service) ListenUDP() (net.PacketConn, error) {
	return s.adapter.ListenUDP(s.ID(), s.Name(), s.BindPort(), s.Address())
}

func (s *service) Dial() (*grpc.ClientConn, error) {
	return s.adapter.DialService(s.name, s.tag)
}
func (s *service) DiscoverEndpoints() ([]*pb.NodeService, error) {
	return s.adapter.EndpointsByService(s.name, s.tag)
}
func (s *service) Address() string {
	return s.address
}
func (s *service) AdvertisedHost() string {
	host, _, err := net.SplitHostPort(s.address)
	if err != nil {
		panic(err)
	}
	return host
}
func (s *service) AdvertisedPort() int {
	_, port, err := net.SplitHostPort(s.address)
	if err != nil {
		panic(err)
	}
	portInt, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		panic(err)
	}
	return int(portInt)
}
func (s *service) ID() string {
	return s.id
}
func (s *service) Name() string {
	return s.name
}
func (s *service) BindPort() int {
	return s.bindPort
}

func NewService(id, name, tag, address string, bindPort int, adapter DiscoveryAdapter) Service {
	return &service{
		id:       id,
		tag:      tag,
		name:     name,
		address:  address,
		adapter:  adapter,
		bindPort: bindPort,
	}
}

func NewServiceFromIdentity(id identity.Identity, adapter DiscoveryAdapter) Service {
	if id == nil {
		panic("nil identity")
	}
	return &service{
		id:       id.ID(),
		tag:      id.Tag(),
		name:     id.Name(),
		address:  fmt.Sprintf("%s:%d", id.AdvertisedAddress(), id.AdvertisedPort()),
		adapter:  adapter,
		bindPort: id.BindPort(),
	}
}
