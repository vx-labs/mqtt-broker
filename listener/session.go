package listener

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/vx-labs/mqtt-broker/pool"
	"github.com/vx-labs/mqtt-broker/struct/queues/inflight"
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
	logger := local.logger.WithOptions(zap.Fields(fields...))
	logger.Info("accepted new connection")
	session := &localSession{
		encoder:   encoder.New(t.Channel),
		transport: t.Channel,
		quit:      make(chan struct{}),
	}
	timer := CONNECT_DEADLINE
	enc := encoder.New(t.Channel)
	inflight := inflight.New(enc.Publish)
	defer close(session.quit)
	publishWorkers := pool.NewPool(5)
	dec := decoder.New(
		decoder.OnConnect(func(p *packet.Connect) error {
			id, token, connack, err := local.broker.Connect(ctx, t, p)
			if err != nil {
				logger.Error("CONNECT failed", zap.Error(err))
				enc.ConnAck(connack)
				return ErrConnectNotDone
			}
			if connack.ReturnCode != packet.CONNACK_CONNECTION_ACCEPTED {
				enc.ConnAck(connack)
				return ErrConnectNotDone
			}
			if id == "" {
				logger.Error("broker returned an empty session id")
				enc.ConnAck(&packet.ConnAck{
					ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
					Header:     &packet.Header{},
				})
				return ErrConnectNotDone
			}
			if token == "" {
				logger.Error("broker returned an empty session token")
				enc.ConnAck(&packet.ConnAck{
					ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
					Header:     &packet.Header{},
				})
				return ErrConnectNotDone
			}
			logger = logger.WithOptions(zap.Fields(zap.String("session_id", id), zap.String("client_id", string(p.ClientId))))
			session.id = id
			session.token = token
			local.mutex.Lock()
			old := local.sessions.ReplaceOrInsert(session)
			local.mutex.Unlock()
			if old != nil {
				old.(*localSession).transport.Close()
			}

			go func() {
				ticker := time.NewTicker(200 * time.Millisecond)
				defer ticker.Stop()
				var offset uint64 = 0
				var messages []*packet.Publish
				var err error
				logger.Debug("started queue poller")
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						backoff.Retry(func() error {
							select {
							case <-ctx.Done():
								return nil
							default:
							}
							offset, messages, err = local.queues.GetMessages(ctx, session.id, offset)
							if err != nil {
								logger.Error("failed to poll for messages", zap.Error(err))
								return err
							}
							for _, message := range messages {
								inflight.Put(message)
							}
							return nil
						}, backoff.NewExponentialBackOff())
					}
				}
			}()
			enc.ConnAck(&packet.ConnAck{
				ReturnCode: packet.CONNACK_CONNECTION_ACCEPTED,
				Header:     &packet.Header{},
			})
			timer = p.KeepaliveTimer
			renewDeadline(timer, t.Channel)
			logger.Info("started session")
			return nil
		}),
		decoder.OnPublish(func(p *packet.Publish) error {
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			return publishWorkers.Call(func() error {
				puback, err := local.broker.Publish(ctx, session.token, p)
				if err != nil {
					logger.Error("failed to publish message", zap.Error(err))
					return err
				}
				if puback != nil {
					return enc.PubAck(puback)
				}
				return nil
			})
		}),
		decoder.OnSubscribe(func(p *packet.Subscribe) error {
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			suback, err := local.broker.Subscribe(ctx, session.token, p)
			if err != nil {
				return err
			}
			return enc.SubAck(suback)
		}),
		decoder.OnUnsubscribe(func(p *packet.Unsubscribe) error {
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			unsuback, err := local.broker.Unsubscribe(ctx, session.token, p)
			if err != nil {
				return err
			}
			return enc.UnsubAck(unsuback)
		}),
		decoder.OnPubAck(func(p *packet.PubAck) error {
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			inflight.Ack(p)
			return nil
		}),
		decoder.OnPingReq(func(p *packet.PingReq) error {
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			pingresp, err := local.broker.PingReq(ctx, session.token, p)
			if err != nil {
				return err
			}
			return session.encoder.PingResp(pingresp)
		}),
		decoder.OnDisconnect(func(p *packet.Disconnect) error {
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			local.broker.Disconnect(ctx, session.token, p)
			return ErrSessionDisconnected
		}),
	)
	renewDeadline(CONNECT_DEADLINE, t.Channel)
	var err error
	for {
		err = dec.Decode(t.Channel)
		if err != nil {
			if err == ErrConnectNotDone {
				break
			}
			if err == ErrSessionDisconnected {
				logger.Info("session disconnected")
				break
			}
			logger.Info("session lost", zap.Error(err))
			break
		}
	}
	t.Channel.Close()
	if session.id != "" {
		err = local.broker.CloseSession(ctx, session.token)
		if err != nil {
			logger.Warn("failed to close session on broker", zap.Error(err))
		}
	}
	local.mutex.Lock()
	local.sessions.Delete(session)
	local.mutex.Unlock()
	inflight.Close()
}
