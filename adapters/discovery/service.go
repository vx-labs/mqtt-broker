package discovery

import (
	"fmt"
	"net"
	"strconv"
	"time"

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

type serviceTCPListener struct {
	closeCallback func() error
	listener      net.Listener
}

func (s *serviceTCPListener) Accept() (net.Conn, error) {
	return s.listener.Accept()
}
func (s *serviceTCPListener) Addr() net.Addr {
	return s.listener.Addr()
}
func (s *serviceTCPListener) Close() error {
	err := s.closeCallback()
	if err != nil {
		return err
	}
	return s.listener.Close()
}

type serviceUDPListener struct {
	closeCallback func() error
	listener      net.PacketConn
}

func (s *serviceUDPListener) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return s.listener.ReadFrom(p)
}
func (s *serviceUDPListener) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return s.listener.WriteTo(p, addr)
}
func (s *serviceUDPListener) LocalAddr() net.Addr {
	return s.listener.LocalAddr()
}
func (s *serviceUDPListener) SetDeadline(t time.Time) error {
	return s.listener.SetDeadline(t)
}
func (s *serviceUDPListener) SetReadDeadline(t time.Time) error {
	return s.listener.SetReadDeadline(t)
}
func (s *serviceUDPListener) SetWriteDeadline(t time.Time) error {
	return s.listener.SetWriteDeadline(t)
}
func (s *serviceUDPListener) Close() error {
	err := s.closeCallback()
	if err != nil {
		return err
	}
	return s.listener.Close()
}

func (s *service) ListenTCP() (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.BindPort()))
	if err != nil {
		return nil, err
	}
	err = s.RegisterTCP()
	if err != nil {
		return nil, err
	}
	return &serviceTCPListener{
		closeCallback: s.Unregister,
		listener:      listener,
	}, nil
}
func (s *service) ListenUDP() (net.PacketConn, error) {
	listener, err := net.ListenPacket("udp", fmt.Sprintf(":%d", s.BindPort()))
	if err != nil {
		return nil, err
	}
	/*err = s.RegisterUDP()
	if err != nil {
		return nil, err
	}*/
	return &serviceUDPListener{
		//	closeCallback: s.Unregister,
		listener: listener,
	}, nil
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
func (s *service) RegisterTCP() error {
	return s.adapter.RegisterTCPService(s.id, s.name, s.address)
}
func (s *service) RegisterUDP() error {
	return s.adapter.RegisterUDPService(s.id, s.name, s.address)
}
func (s *service) RegisterGRPC() error {
	return s.adapter.RegisterGRPCService(s.id, s.name, s.address)
}
func (s *service) Unregister() error {
	return s.adapter.UnregisterService(s.id)
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
