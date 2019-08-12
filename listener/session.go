package listener

import (
	"context"
	"errors"
	"io"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/vx-labs/mqtt-broker/pool"
	"github.com/vx-labs/mqtt-broker/queues/inflight"
	publishQueue "github.com/vx-labs/mqtt-broker/queues/publish"
	"github.com/vx-labs/mqtt-broker/transport"
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
	fields := []zapcore.Field{
		zap.String("remote_address", t.RemoteAddress),
		zap.String("transport", t.Name),
	}
	local.logger.Info("accepted new connection", fields...)
	session := &localSession{
		encoder:   encoder.New(t.Channel),
		transport: t.Channel,
		quit:      make(chan struct{}),
	}
	timer := CONNECT_DEADLINE
	enc := encoder.New(t.Channel)
	inflight := inflight.New(enc.Publish)
	queue := publishQueue.New()
	defer close(session.quit)
	session.queue = queue
	publishWorkers := pool.NewPool(5)
	dec := decoder.New(
		decoder.OnConnect(func(p *packet.Connect) error {
			id, connack, err := local.broker.Connect(ctx, t, p)
			if err != nil {
				local.logger.Error("failed to send CONNECT", append(fields, zap.Error(err))...)
				return enc.ConnAck(connack)
			}
			if id == "" {
				local.logger.Error("broker returned an empty session id", fields...)
				enc.ConnAck(&packet.ConnAck{
					ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
					Header:     &packet.Header{},
				})
				return errors.New("broker returned an empty session id")
			}
			fields = append(fields, zap.String("session_id", id), zap.String("client_id", string(p.ClientId)))
			session.id = id
			local.mutex.Lock()
			old := local.sessions.ReplaceOrInsert(session)
			local.mutex.Unlock()
			if old != nil {
				old.(*localSession).transport.Close()
			}
			enc.ConnAck(&packet.ConnAck{
				ReturnCode: packet.CONNACK_CONNECTION_ACCEPTED,
				Header:     &packet.Header{},
			})
			timer = p.KeepaliveTimer
			renewDeadline(timer, t.Channel)
			local.logger.Info("started session", fields...)
			return nil
		}),
		decoder.OnPublish(func(p *packet.Publish) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			return publishWorkers.Call(func() error {
				puback, err := local.broker.Publish(ctx, session.id, p)
				if err != nil {
					local.logger.Error("failed to publish message", append(fields, zap.Error(err))...)
					return err
				}
				if puback != nil {
					return enc.PubAck(puback)
				}
				return nil
			})
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
			inflight.Ack(p)
			return nil
		}),
		decoder.OnPingReq(func(p *packet.PingReq) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			pingresp, err := local.broker.PingReq(ctx, session.id, p)
			if err != nil {
				return err
			}
			return session.encoder.PingResp(pingresp)
		}),
		decoder.OnDisconnect(func(p *packet.Disconnect) error {
			if session.id == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			local.broker.Disconnect(ctx, session.id, p)
			return ErrSessionDisconnected
		}),
	)
	renewDeadline(CONNECT_DEADLINE, t.Channel)
	go queue.Consume(func(p *publishQueue.Message) {
		inflight.Put(p.Publish)
	})

	var err error
	for {
		err = dec.Decode(t.Channel)
		if err != nil {
			if err == ErrSessionDisconnected {
				local.logger.Info("session disconnected", fields...)
				break
			}
			local.logger.Warn("session lost", fields...)
			if err != io.EOF {
				local.logger.Error("decoding failed", append(fields, zap.Error(err))...)
			}
			break
		}
	}
	err = t.Channel.Close()
	if err != nil {
		local.logger.Warn("failed to close session channel", append(fields, zap.Error(err))...)
	}
	err = local.broker.CloseSession(ctx, session.id)
	if err != nil {
		local.logger.Warn("failed to close session on broker", append(fields, zap.Error(err))...)
	}
	local.mutex.Lock()
	local.sessions.Delete(session)
	local.mutex.Unlock()
	queue.Close()
	inflight.Close()
}
