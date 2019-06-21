package broker

import (
	"errors"
	"time"

	"github.com/vx-labs/mqtt-protocol/encoder"

	"github.com/vx-labs/mqtt-broker/broker/transport"
	"github.com/vx-labs/mqtt-broker/queues/inflight"
	"github.com/vx-labs/mqtt-broker/queues/messages"
	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrSessionDisconnected = errors.New("session disconnected")
)

type Session struct {
	id        string
	tenant    string
	keepalive int32
	closed    bool
	transport transport.Metadata
	connect   *packet.Connect
	encoder   *encoder.Encoder
	queue     *messages.Queue
	inflight  *inflight.Queue
	quit      chan struct{}
}

func newSession(transport transport.Metadata, queueSize int) *Session {
	s := &Session{
		keepalive: 30,
		transport: transport,
		quit:      make(chan struct{}),
	}
	go func() {
		<-s.quit
		s.queue.Close()
	}()

	return s
}

func (s *Session) TransportName() string {
	return s.transport.Name
}
func (s *Session) RemoteAddress() string {
	return s.transport.RemoteAddress
}
func (s *Session) Close() error {
	s.closed = true
	return s.transport.Channel.Close()
}
func (s *Session) Connect() *packet.Connect {
	return s.connect
}
func (s *Session) ID() string {
	return s.id
}
func (s *Session) Tenant() string {
	return s.tenant
}

func (s *Session) Publish(p *packet.Publish) error {
	s.queue.Put(time.Now(), p)
	return nil
}

func (s *Session) SubAck(mid int32, grantedQoS []int32) error {
	return s.encoder.SubAck(&packet.SubAck{
		Header:    &packet.Header{},
		MessageId: mid,
		Qos:       grantedQoS,
	})
}
func (s *Session) PubAck(mid int32) error {
	return s.encoder.PubAck(&packet.PubAck{
		Header:    &packet.Header{},
		MessageId: mid,
	})
}

func (s *Session) ConnAck(returnCode int32) error {
	return s.encoder.ConnAck(&packet.ConnAck{
		Header:     &packet.Header{},
		ReturnCode: returnCode,
	})
}

func (s *Session) RenewDeadline() {
	conn := s.transport.Channel
	deadline := time.Now().Add(time.Duration(s.keepalive) * time.Second * 2)
	conn.SetDeadline(deadline)
}
