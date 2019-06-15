package listener

import (
	"errors"
	"time"

	"github.com/vx-labs/mqtt-protocol/encoder"

	"github.com/vx-labs/mqtt-broker/broker/listener/transport"
	"github.com/vx-labs/mqtt-broker/events"
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
	events    *events.Bus
	queue     *messages.Queue
	inflight  *inflight.Queue
	quit      chan struct{}
}

func newSession(transport transport.Metadata, queueSize int) *Session {
	s := &Session{
		events:    events.NewEventBus(),
		keepalive: 30,
		transport: transport,
		quit:      make(chan struct{}),
	}
	go func() {
		<-s.quit
		s.queue.Close()
		s.events.Close()
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

func (s *Session) emitPublish(packet *packet.Publish) {
	s.events.Emit(events.Event{
		Key:   "session_published",
		Entry: packet,
	})
}
func (s *Session) Publish(p *packet.Publish) error {
	s.queue.Put(time.Now(), p)
	return nil
}
func (s *Session) OnPublish(f func(packet *packet.Publish)) func() {
	return s.events.Subscribe("session_published", func(event events.Event) {
		payload := event.Entry.(*packet.Publish)
		f(payload)
	})
}
func (s *Session) emitSubscribe(packet *packet.Subscribe) {
	s.events.Emit(events.Event{
		Key:   "session_subscribed",
		Entry: packet,
	})
}

func (s *Session) OnSubscribe(f func(packet *packet.Subscribe)) func() {
	return s.events.Subscribe("session_subscribed", func(event events.Event) {
		payload := event.Entry.(*packet.Subscribe)
		f(payload)
	})
}
func (s *Session) emitUnsubscribe(packet *packet.Unsubscribe) {
	s.events.Emit(events.Event{
		Key:   "session_unsubscribed",
		Entry: packet,
	})
}

func (s *Session) OnUnsubscribe(f func(packet *packet.Unsubscribe)) func() {
	return s.events.Subscribe("session_unsubscribed", func(event events.Event) {
		payload := event.Entry.(*packet.Unsubscribe)
		f(payload)
	})
}
func (s *Session) emitClosed() {
	s.events.Emit(events.Event{
		Key:   "session_closed",
		Entry: nil,
	})
}

func (s *Session) OnClosed(f func()) func() {
	return s.events.Subscribe("session_closed", func(_ events.Event) {
		f()
	})
}
func (s *Session) emitLost() {
	s.events.Emit(events.Event{
		Key:   "session_lost",
		Entry: nil,
	})
}

func (s *Session) OnLost(f func()) func() {
	return s.events.Subscribe("session_lost", func(_ events.Event) {
		f()
	})
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

func (s *Session) emitConnected(p *packet.Connect) {
	s.events.Emit(events.Event{
		Key:   "session_connected",
		Entry: p,
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
