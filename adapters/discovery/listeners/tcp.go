package listeners

import (
	"net"
	"time"
)

type ServiceTCPListener struct {
	CloseCallback func() error
	Listener      net.Listener
}

func (s *ServiceTCPListener) Accept() (net.Conn, error) {
	return s.Listener.Accept()
}
func (s *ServiceTCPListener) Addr() net.Addr {
	return s.Listener.Addr()
}
func (s *ServiceTCPListener) Close() error {
	if s.CloseCallback != nil {
		s.CloseCallback()
	}
	return s.Listener.Close()
}

type ServiceUDPListener struct {
	CloseCallback func() error
	Listener      net.PacketConn
}

func (s *ServiceUDPListener) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return s.Listener.ReadFrom(p)
}
func (s *ServiceUDPListener) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return s.Listener.WriteTo(p, addr)
}
func (s *ServiceUDPListener) LocalAddr() net.Addr {
	return s.Listener.LocalAddr()
}
func (s *ServiceUDPListener) SetDeadline(t time.Time) error {
	return s.Listener.SetDeadline(t)
}
func (s *ServiceUDPListener) SetReadDeadline(t time.Time) error {
	return s.Listener.SetReadDeadline(t)
}
func (s *ServiceUDPListener) SetWriteDeadline(t time.Time) error {
	return s.Listener.SetWriteDeadline(t)
}
func (s *ServiceUDPListener) Close() error {
	if s.CloseCallback != nil {
		s.CloseCallback()
	}
	return s.Listener.Close()
}
