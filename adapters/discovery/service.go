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
	adapter  DiscoveryAdapter
	id       string
	name     string
	address  string
	bindPort int
}

func (s *service) ListenTCP() (net.Listener, error) {
	return s.adapter.ListenTCP(s.ID(), s.Name(), s.BindPort(), s.Address())
}
func (s *service) ListenUDP() (net.PacketConn, error) {
	return s.adapter.ListenUDP(s.ID(), s.Name(), s.BindPort(), s.Address())
}

func (s *service) Dial(tags ...string) (*grpc.ClientConn, error) {
	return s.adapter.DialService(s.name, tags...)
}
func (s *service) DiscoverEndpoints() ([]*pb.NodeService, error) {
	return s.adapter.EndpointsByService(s.name)
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
func (s *service) AddTag(key string, value string) error {
	return s.adapter.AddServiceTag(s.id, key, value)
}
func (s *service) RemoveTag(key string) error {
	return s.adapter.RemoveServiceTag(s.id, key)
}

func NewService(id, name, address string, bindPort int, adapter DiscoveryAdapter) Service {
	return &service{
		id:       id,
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
		name:     id.Name(),
		address:  fmt.Sprintf("%s:%d", id.AdvertisedAddress(), id.AdvertisedPort()),
		adapter:  adapter,
		bindPort: id.BindPort(),
	}
}
