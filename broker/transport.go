package broker

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/vx-labs/mqtt-broker/sessions"

	"github.com/vx-labs/mqtt-broker/broker/transport"
	"github.com/vx-labs/mqtt-broker/queues/inflight"
	"github.com/vx-labs/mqtt-broker/queues/messages"

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
type listener struct {
	ch chan transport.Metadata
}

func (l *listener) Close() error {
	close(l.ch)
	return nil
}

func (b *Broker) NewListener(inflightSize int) (io.Closer, chan<- transport.Metadata) {
	ch := make(chan transport.Metadata)

	l := &listener{
		ch: ch,
	}
	go func() {
		for transport := range ch {
			go b.runSession(transport, inflightSize)
		}
	}()
	return l, l.ch
}

func validateClientID(clientID []byte) bool {
	return len(clientID) > 0 && len(clientID) < 128
}

func (handler *Broker) runSession(t transport.Metadata, inflightSize int) {
	log.Printf("INFO: listener %s: accepted new connection from %s", t.Name, t.RemoteAddress)
	c := t.Channel
	session := newSession(t, inflightSize)
	session.encoder = encoder.New(c)
	session.inflight = inflight.New(session.encoder.Publish)
	session.queue = messages.NewQueue()
	var sessionMetadata sessions.Session
	defer close(session.quit)
	dec := decoder.New(
		decoder.OnConnect(func(p *packet.Connect) error {
			return handler.workers.Call(func() error {
				session.keepalive = p.KeepaliveTimer
				session.renewDeadline()
				clientID := p.ClientId
				if !validateClientID(clientID) {
					session.ConnAck(packet.CONNACK_REFUSED_IDENTIFIER_REJECTED)
					return fmt.Errorf("invalid client ID")
				}
				tenant, id, err := handler.Authenticate(t, clientID, string(p.Username), string(p.Password))
				if err != nil {
					log.Printf("WARN: session %s failed authentication: %v", session.id, err)
					session.encoder.ConnAck(&packet.ConnAck{
						Header:     p.Header,
						ReturnCode: packet.CONNACK_REFUSED_BAD_USERNAME_OR_PASSWORD,
					})
					session.ConnAck(packet.CONNACK_REFUSED_BAD_USERNAME_OR_PASSWORD)
					return fmt.Errorf("authentication failed")
				}
				session.id = id
				session.tenant = tenant
				log.Printf("INFO: starting session %s", session.id)
				var code int32
				sessionMetadata, code, err = handler.OnConnect(session, p)
				if err == nil {
					session.renewDeadline()
				} else {
					log.Printf("ERROR: session %s start failed: %v", session.id, err)
				}
				return session.ConnAck(code)
			})
		}),
		decoder.OnPublish(func(p *packet.Publish) error {
			session.renewDeadline()
			return handler.publishPool.Call(func() error {
				err := handler.OnPublish(sessionMetadata, p)
				if p.Header.Qos == 0 {
					return nil
				}
				if err != nil {
					log.Printf("ERR: failed to handle message publish: %v", err)
					return err
				}
				if p.Header.Qos == 1 {
					return session.PubAck(p.MessageId)
				}
				return nil
			})
		}),
		decoder.OnSubscribe(func(p *packet.Subscribe) error {
			session.renewDeadline()
			return handler.workers.Call(func() error {
				err := handler.OnSubscribe(session, sessionMetadata, p)
				if err == nil {
					qos := make([]int32, len(p.Qos))

					// QoS2 is not supported for now
					for idx := range p.Qos {
						if p.Qos[idx] > 1 {
							qos[idx] = 1
						} else {
							qos[idx] = p.Qos[idx]
						}
					}
					session.SubAck(p.MessageId, qos)
				}
				return nil
			})
		}),
		decoder.OnUnsubscribe(func(p *packet.Unsubscribe) error {
			session.renewDeadline()
			return handler.workers.Call(func() error {
				return handler.OnUnsubscribe(sessionMetadata, p)
			})
		}),
		decoder.OnPubAck(func(p *packet.PubAck) error {
			session.renewDeadline()
			session.inflight.Ack(p.MessageId)
			return nil
		}),
		decoder.OnPingReq(func(p *packet.PingReq) error {
			session.renewDeadline()
			return session.encoder.PingResp(&packet.PingResp{
				Header: p.Header,
			})
		}),
		decoder.OnDisconnect(func(p *packet.Disconnect) error {
			session.renewDeadline()
			return ErrSessionDisconnected
		}),
	)
	c.SetDeadline(
		time.Now().Add(15 * time.Second),
	)
	go func() {
		for {
			select {
			case <-session.quit:
				return
			default:
				msg, err := session.queue.Pop()
				if err == nil {
					session.inflight.Put(msg)
				}
			}
		}
	}()

	var err error
	for {
		err = dec.Decode(c)
		if err != nil {
			if err == io.EOF || err == ErrSessionDisconnected || session.closed {
				break
			}
			if opErr, ok := err.(*net.OpError); ok {
				if opErr.Timeout() {
					log.Printf("ERR: listener %s: session_id=%q read timeout", t.Name, session.id)
					break
				}
			}
			log.Printf("ERR: listener %s: decoding from session_id=%q failed: %v", t.Name, session.id, err)
			break
		}
	}
	c.Close()

	if err != nil && err != ErrSessionDisconnected && !session.closed {
		log.Printf("WARN: listener %s: session %s lost: %v", t.Name, session.id, err)
		handler.workers.Call(func() error {
			handler.Sessions.Delete(sessionMetadata.ID, "session_lost")
			err := handler.deleteSessionSubscriptions(sessionMetadata)
			if err != nil {
				log.Printf("WARN: failed to delete session subscriptions: %v", err)
			}
			if len(sessionMetadata.WillTopic) > 0 {
				handler.OnPublish(sessionMetadata, &packet.Publish{
					Header: &packet.Header{
						Dup:    false,
						Retain: sessionMetadata.WillRetain,
						Qos:    sessionMetadata.WillQoS,
					},
					Payload: sessionMetadata.WillPayload,
					Topic:   sessionMetadata.WillTopic,
				})
			}
			return nil
		})
	} else {
		log.Printf("INFO: listener %s: session %s closed", t.Name, session.id)
		handler.workers.Call(func() error {
			handler.Sessions.Delete(sessionMetadata.ID, "session_disconnected")
			err := handler.deleteSessionSubscriptions(sessionMetadata)
			if err != nil {
				log.Printf("WARN: failed to delete session subscriptions: %v", err)
				return err
			}
			return nil
		})
	}
}
