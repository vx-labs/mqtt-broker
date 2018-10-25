package listener

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/vx-labs/mqtt-protocol/decoder"
	"github.com/vx-labs/mqtt-protocol/encoder"
	"github.com/vx-labs/mqtt-protocol/packet"
)

type TimeoutReadWriteCloser interface {
	SetDeadline(time.Time) error
	io.ReadWriteCloser
}
type Transport interface {
	io.Closer
	Name() string
	Encrypted() bool
	EncryptionState() *tls.ConnectionState
	Channel() TimeoutReadWriteCloser
	RemoteAddress() string
}
type Handler interface {
	Authenticate(transport Transport, sessionID, username string, password string) (tenant string, id string, err error)
	OnConnect(id, tenant string, sess *Session)
	OnSubscribe(id string, tenant string, packet *packet.Subscribe) error
	OnUnsubscribe(id string, tenant string, packet *packet.Unsubscribe) error
	OnSessionClosed(id, tenant string)
	OnSessionLost(id, tenant string)
	OnPublish(id, tenant string, packet *packet.Publish) error
}

type listener struct {
	ch      chan Transport
	handler Handler
}

func (l *listener) Close() error {
	close(l.ch)
	return nil
}

func New(handler Handler) (io.Closer, chan<- Transport) {
	ch := make(chan Transport)

	l := &listener{
		ch:      ch,
		handler: handler,
	}
	go func() {
		for transport := range ch {
			go l.runSession(transport)
		}
	}()
	return l, l.ch
}

func (l *listener) runSession(t Transport) {
	log.Printf("INFO: listener %s: accepted new connection from %s", t.Name(), t.RemoteAddress())
	c := t.Channel()
	handler := l.handler
	enc := encoder.New(c)
	session := &Session{
		ch:        make(chan *packet.Publish, 100),
		tenant:    "_default",
		keepalive: 30,
		closer:    t.Close,
	}
	defer close(session.ch)

	dec := decoder.New(
		decoder.OnConnect(func(p *packet.Connect) error {
			session.keepalive = p.KeepaliveTimer
			c.SetDeadline(
				time.Now().Add(time.Duration(session.keepalive) * time.Second * 2),
			)
			clientID := string(p.ClientId)
			tenant, id, err := handler.Authenticate(t, clientID, string(p.Username), string(p.Password))
			if err != nil {
				enc.ConnAck(&packet.ConnAck{
					Header:     p.Header,
					ReturnCode: packet.CONNACK_REFUSED_BAD_USERNAME_OR_PASSWORD,
				})
				return fmt.Errorf("authentication failed")
			}
			session.id = id
			session.connect = p
			session.tenant = tenant
			handler.OnConnect(session.id, tenant, session)
			log.Printf("INFO: listener %s: session %s started", t.Name(), session.id)
			return enc.ConnAck(&packet.ConnAck{
				Header:     p.Header,
				ReturnCode: packet.CONNACK_CONNECTION_ACCEPTED,
			})
		}),
		decoder.OnPublish(func(p *packet.Publish) error {
			c.SetDeadline(
				time.Now().Add(time.Duration(session.keepalive) * time.Second * 2),
			)
			err := handler.OnPublish(session.id, session.tenant, p)
			if p.Header.Qos == 0 {
				return nil
			}
			if err != nil {
				log.Printf("ERR: failed to handle message publish: %v", err)
				return err
			}
			if p.Header.Qos == 1 {
				return enc.PubAck(&packet.PubAck{
					Header:    &packet.Header{},
					MessageId: p.MessageId,
				})
			}
			return nil
		}),
		decoder.OnSubscribe(func(p *packet.Subscribe) error {
			c.SetDeadline(
				time.Now().Add(time.Duration(session.keepalive) * time.Second * 2),
			)
			err := handler.OnSubscribe(session.id, session.tenant, p)
			if err != nil {
				return err
			}
			qos := make([]int32, len(p.Qos))

			// QoS2 is not supported for now
			for idx := range p.Qos {
				if p.Qos[idx] > 1 {
					qos[idx] = 1
				} else {
					qos[idx] = p.Qos[idx]
				}
			}
			return enc.SubAck(&packet.SubAck{
				Header:    p.Header,
				MessageId: p.MessageId,
				Qos:       qos,
			})
		}),
		decoder.OnUnsubscribe(func(p *packet.Unsubscribe) error {
			return handler.OnUnsubscribe(session.id, session.tenant, p)
		}),
		decoder.OnPubAck(func(*packet.PubAck) error { return nil }),
		decoder.OnPingReq(func(p *packet.PingReq) error {
			c.SetDeadline(
				time.Now().Add(time.Duration(session.keepalive) * time.Second * 2),
			)
			return enc.PingResp(&packet.PingResp{
				Header: p.Header,
			})
		}),
		decoder.OnDisconnect(func(p *packet.Disconnect) error {
			return ErrSessionDisconnected
		}),
	)
	c.SetDeadline(
		time.Now().Add(10 * time.Second),
	)
	decoderCh := make(chan struct{})

	go func() {
		for p := range session.ch {
			err := enc.Publish(p)
			if err != nil {
				log.Printf("ERR: failed to send message to %s: %v", session.id, err)
			}
		}
	}()
	var err error
	go func() {
		defer close(decoderCh)
		for {
			err = dec.Decode(c)
			if err == io.EOF || err == ErrSessionDisconnected {
				return
			}
			if err != nil {
				log.Printf("ERR: listener %s: decoding failed: %v", t.Name(), err)
				return
			}
		}
	}()

	select {
	case <-decoderCh:
	}
	c.Close()
	if err != nil && err != ErrSessionDisconnected {
		log.Printf("WARN: listener %s: session %s lost: %v", t.Name(), session.id, err)
		handler.OnSessionLost(session.id, session.tenant)
	} else {
		log.Printf("INFO: listener %s: session %s closed", t.Name(), session.id)
		handler.OnSessionClosed(session.id, session.tenant)
	}
}
