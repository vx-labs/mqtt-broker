package listener

import (
	"errors"

	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrSessionDisconnected = errors.New("session disconnected")
)

type Session struct {
	id        string
	tenant    string
	keepalive int32
	connect   *packet.Connect
	ch        chan *packet.Publish
	closer    func() error
}

func (s *Session) Close() error {
	return s.closer()
}
func (s *Session) Connect() *packet.Connect {
	return s.connect
}
func (s *Session) Channel() chan<- *packet.Publish {
	return s.ch
}
