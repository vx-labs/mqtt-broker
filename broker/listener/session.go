package listener

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vx-labs/mqtt-broker/broker/transport"
	"github.com/vx-labs/mqtt-broker/queues/inflight"
	"github.com/vx-labs/mqtt-broker/queues/messages"
	"github.com/vx-labs/mqtt-protocol/decoder"
	"github.com/vx-labs/mqtt-protocol/encoder"
	"github.com/vx-labs/mqtt-protocol/packet"
)

var CONNECT_DEADLINE int32 = 15

var (
	ErrSessionDisconnected = errors.New("session disconnected")
	ErrConnectNotDone      = errors.New("CONNECT not done")
)

func renewDeadline(timer int32, conn transport.TimeoutReadWriteCloser) {
	deadline := time.Now().Add(time.Duration(timer) * time.Second * 2)
	conn.SetDeadline(deadline)
}

func (local *endpoint) runLocalSession(t transport.Metadata) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logrus.New().WithField("service", "listener").WithField("listener", t.Name).WithField("remote", t.RemoteAddress)
	logger.Info("accepted new connection")
	session := &localSession{
		encoder:   encoder.New(t.Channel),
		transport: t.Channel,
		quit:      make(chan struct{}),
	}
	timer := CONNECT_DEADLINE
	enc := encoder.New(t.Channel)
	inflight := inflight.New(enc.Publish)
	queue := messages.NewQueue()
	defer close(session.quit)

	dec := decoder.New(
		decoder.OnConnect(func(p *packet.Connect) error {
			id, connack, err := local.broker.Connect(ctx, t, p)
			if err != nil {
				return enc.ConnAck(connack)
			}
			if id == "" {
				logger.Error("broker returned an empty session id")
				enc.ConnAck(&packet.ConnAck{
					ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
					Header:     &packet.Header{},
				})
				return errors.New("broker returned an empty session id")
			}
			logger = logger.WithField("session_id", id)
			logger = logger.WithField("session_client_id", string(p.ClientId))
			session.id = id
			local.mutex.Lock()
			old := local.sessions.ReplaceOrInsert(session)
			local.mutex.Unlock()
			if old != nil {
				old.(*localSession).transport.Close()
			}
			logger.Info("starting session")
			enc.ConnAck(&packet.ConnAck{
				ReturnCode: packet.CONNACK_CONNECTION_ACCEPTED,
				Header:     &packet.Header{},
			})
			timer = p.KeepaliveTimer
			renewDeadline(timer, t.Channel)
			return nil
		}),
		decoder.OnPublish(func(p *packet.Publish) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			puback, err := local.broker.Publish(ctx, session.id, p)
			if err != nil {
				return err
			}
			if puback != nil {
				return enc.PubAck(puback)
			}
			return nil
		}),
		decoder.OnSubscribe(func(p *packet.Subscribe) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			suback, err := local.broker.Subscribe(ctx, session.id, p)
			if err != nil {
				return err
			}
			return enc.SubAck(suback)
		}),
		decoder.OnUnsubscribe(func(p *packet.Unsubscribe) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			unsuback, err := local.broker.Unsubscribe(ctx, session.id, p)
			if err != nil {
				return err
			}
			return enc.UnsubAck(unsuback)
		}),
		decoder.OnPubAck(func(p *packet.PubAck) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			inflight.Ack(p.MessageId)
			return nil
		}),
		decoder.OnPingReq(func(p *packet.PingReq) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			return session.encoder.PingResp(&packet.PingResp{
				Header: p.Header,
			})
		}),
		decoder.OnDisconnect(func(p *packet.Disconnect) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			return ErrSessionDisconnected
		}),
	)
	renewDeadline(CONNECT_DEADLINE, t.Channel)
	go func() {
		for {
			select {
			case <-session.quit:
				return
			default:
				msg, err := queue.Pop()
				if err == nil {
					inflight.Put(msg)
				}
			}
		}
	}()

	var err error
	for {
		err = dec.Decode(t.Channel)
		if err != nil {
			if err == io.EOF || err == ErrSessionDisconnected {
				logger.Info("session disconnected")
				break
			}
			if opErr, ok := err.(*net.OpError); ok {
				if opErr.Timeout() {
					logger.Errorf("read timeout")
					break
				}
			}
			logger.Errorf("decoding failed: %v", err)
			break
		}
	}
	t.Channel.Close()
	local.broker.CloseSession(ctx, session.id)
	local.mutex.Lock()
	local.sessions.Delete(session)
	local.mutex.Unlock()
}
