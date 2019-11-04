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
	session := &localSession{
		encoder:   encoder.New(t.Channel),
		transport: t.Channel,
		logger:    logger,
	}
	defer func() {
		t.Channel.Close()
		local.mutex.Lock()
		local.sessions.Delete(session)
		local.mutex.Unlock()
		cancel()
	}()
	err := local.handleSessionPackets(ctx, session, t)
	if err != nil {
		if session.id != "" {
			local.broker.CloseSession(ctx, session.token)
		}
	}
}
func (local *endpoint) handleSessionPackets(ctx context.Context, session *localSession, t transport.Metadata) error {
	session.logger.Info("accepted new connection")
	timer := CONNECT_DEADLINE
	enc := encoder.New(t.Channel)
	inflight := inflight.New(enc.Publish)
	publishWorkers := pool.NewPool(5)
	dec := decoder.Async(t.Channel)
	renewDeadline(CONNECT_DEADLINE, t.Channel)
	defer func() {
		inflight.Close()
		dec.Cancel()
	}()
	for data := range dec.Packet() {
		switch p := data.(type) {
		case *packet.Connect:
			id, token, connack, err := local.broker.Connect(ctx, t, p)
			if err != nil {
				session.logger.Error("CONNECT failed", zap.Error(err))
				enc.ConnAck(connack)
				return ErrConnectNotDone
			}
			if connack.ReturnCode != packet.CONNACK_CONNECTION_ACCEPTED {
				enc.ConnAck(connack)
				return ErrConnectNotDone
			}
			if id == "" {
				session.logger.Error("broker returned an empty session id")
				enc.ConnAck(&packet.ConnAck{
					ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
					Header:     &packet.Header{},
				})
				return ErrConnectNotDone
			}
			if token == "" {
				session.logger.Error("broker returned an empty session token")
				enc.ConnAck(&packet.ConnAck{
					ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
					Header:     &packet.Header{},
				})
				return ErrConnectNotDone
			}
			session.logger = session.logger.WithOptions(zap.Fields(zap.String("session_id", id), zap.String("client_id", string(p.ClientId))))
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
				session.logger.Debug("started queue poller")
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						count := 0
						backoff.Retry(func() error {
							select {
							case <-ctx.Done():
								return nil
							default:
							}
							offset, messages, err = local.queues.GetMessages(ctx, session.id, offset)
							if err != nil {
								session.logger.Error("failed to poll for messages", zap.Error(err))
								count++
								if count >= 5 {
									return backoff.Permanent(err)
								}
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
			session.logger.Info("started session")
		case *packet.Publish:
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			err := publishWorkers.Call(func() error {
				puback, err := local.broker.Publish(ctx, session.token, p)
				if err != nil {
					session.logger.Error("failed to publish message", zap.Error(err))
					return err
				}
				if puback != nil {
					return enc.PubAck(puback)
				}
				return nil
			})
			if err != nil {
				return err
			}
		case *packet.Subscribe:
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			suback, err := local.broker.Subscribe(ctx, session.token, p)
			if err != nil {
				return err
			}
			err = enc.SubAck(suback)
			if err != nil {
				return err
			}
		case *packet.Unsubscribe:
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			unsuback, err := local.broker.Unsubscribe(ctx, session.token, p)
			if err != nil {
				return err
			}
			err = enc.UnsubAck(unsuback)
			if err != nil {
				return err
			}
		case *packet.PubAck:
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			inflight.Ack(p)
		case *packet.PingReq:
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			pingresp, err := local.broker.PingReq(ctx, session.token, p)
			if err != nil {
				return err
			}
			err = session.encoder.PingResp(pingresp)
			if err != nil {
				return err
			}
		case *packet.Disconnect:
			if session.token == "" {
				return ErrConnectNotDone
			}
			renewDeadline(timer, t.Channel)
			local.broker.Disconnect(ctx, session.token, p)
			return nil
		}
	}
	return dec.Err()
}
