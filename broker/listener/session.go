package listener

import (
	"errors"

	"github.com/vx-labs/mqtt-protocol/encoder"

	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrSessionDisconnected = errors.New("session disconnected")
)

type Session struct {
	id        string
	tenant    string
	keepalive int32
	transport Transport
	connect   *packet.Connect
	encoder   *encoder.Encoder
	events    *events.Bus
}

func newSession(transport Transport) *Session {
	s := &Session{
		events:    events.NewEventBus(),
		keepalive: 30,
		transport: transport,
	}

	return s
}

func (s *Session) TransportName() string {
	return s.transport.Name()
}
func (s *Session) Close() error {
	return s.transport.Close()
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
	return s.encoder.Publish(p)
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
