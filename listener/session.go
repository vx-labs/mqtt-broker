package listener

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/vx-labs/mqtt-broker/struct/queues/inflight"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/cenkalti/backoff"
	"github.com/vx-labs/mqtt-broker/transport"
	"github.com/vx-labs/mqtt-protocol/decoder"
	"github.com/vx-labs/mqtt-protocol/encoder"
	"github.com/vx-labs/mqtt-protocol/packet"
)

var CONNECT_DEADLINE int32 = 15
var sessionTracePacket = os.Getenv("SESSION_TRACE_PACKET")
var (
	ErrSessionDisconnected = errors.New("session disconnected")
	ErrConnectNotDone      = errors.New("CONNECT not done")
)

type DeadlineSetter interface {
	SetDeadline(time.Time) error
}

func renewDeadline(timer int32, conn DeadlineSetter) error {
	deadline := time.Now().Add(time.Duration(timer) * time.Second * 2)
	return conn.SetDeadline(deadline)
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
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				session.logger.Info("session lost", zap.String("reason", "io timeout"))
			} else {
				session.logger.Info("session lost", zap.String("reason", err.Error()))
			}
			local.broker.CloseSession(ctx, session.token)
		}
	} else {
		session.logger.Info("session disconnected")
	}
}

func (local *endpoint) handleSessionPackets(ctx context.Context, session *localSession, t transport.Metadata) error {
	session.logger.Info("accepted new connection")
	timer := CONNECT_DEADLINE
	enc := encoder.New(t.Channel)
	dec := decoder.Async(t.Channel)
	ingressQueue := inflight.New(enc.Publish)
	defer ingressQueue.Close()

	var poller chan error
	renewDeadline(CONNECT_DEADLINE, t.Channel)
	defer func() {
		dec.Cancel()
	}()
	for data := range dec.Packet() {
		packetCtx, cancelReqCtx := context.WithTimeout(ctx, 15*time.Second)
		if sessionTracePacket == "true" {
			session.logger.Debug("session sent packet", zap.Any("traced_packet", data))
		}
		switch p := data.(type) {
		case *packet.Connect:
			id, token, connack, err := local.broker.Connect(packetCtx, t, p)
			if err != nil {
				session.logger.Error("CONNECT failed", zap.Error(err))
				enc.ConnAck(connack)
				cancelReqCtx()
				return ErrConnectNotDone
			}
			if connack.ReturnCode != packet.CONNACK_CONNECTION_ACCEPTED {
				enc.ConnAck(connack)
				cancelReqCtx()
				return ErrConnectNotDone
			}
			if id == "" {
				session.logger.Error("broker returned an empty session id")
				enc.ConnAck(&packet.ConnAck{
					ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
					Header:     &packet.Header{},
				})
				cancelReqCtx()
				return ErrConnectNotDone
			}
			if token == "" {
				session.logger.Error("broker returned an empty session token")
				enc.ConnAck(&packet.ConnAck{
					ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
					Header:     &packet.Header{},
				})
				cancelReqCtx()
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
			enc.ConnAck(&packet.ConnAck{
				ReturnCode: packet.CONNACK_CONNECTION_ACCEPTED,
				Header:     &packet.Header{},
			})
			timer = p.KeepaliveTimer
			session.logger.Info("started session")
		case *packet.Publish:
			if session.token == "" {
				cancelReqCtx()
				return ErrConnectNotDone
			}
			decodedToken, err := DecodeSessionToken(SigningKey(), session.token)
			if err != nil {
				cancelReqCtx()
				session.logger.Error("failed to validate session token", zap.Error(err))
				return err
			}
			err = local.enqueuePublish(packetCtx, decodedToken.SessionTenant, decodedToken.SessionID, p)
			if err != nil {
				cancelReqCtx()
				session.logger.Error("failed to publish packet", zap.Error(err))
				continue
			}
			switch p.Header.Qos {
			case 1:
				err = enc.PubAck(&packet.PubAck{Header: &packet.Header{}, MessageId: p.MessageId})
				if err != nil {
					cancelReqCtx()
					return err
				}
			default:
				continue
			}
		case *packet.Subscribe:
			if session.token == "" {
				cancelReqCtx()
				return ErrConnectNotDone
			}
			suback, err := local.broker.Subscribe(packetCtx, session.token, p)
			if err != nil {
				cancelReqCtx()
				session.logger.Error("failed to subscribe", zap.Error(err))
				continue
			}
			if poller == nil {
				poller = make(chan error)
				go func() {
					defer close(poller)
					session.logger.Debug("starting queue poller")
					count := 0
					err := backoff.Retry(func() error {
						err := local.queues.StreamMessages(ctx, session.id, func(ackOffset uint64, message *packet.Publish) error {
							if message == nil {
								session.logger.Error("received empty message from queue")
								return nil
							}
							if message.Header == nil {
								message.Header = &packet.Header{Qos: 0}
							}
							switch message.Header.Qos {
							case 0:
								message.MessageId = 1
								enc.Publish(message)
								err = local.queues.AckMessage(ctx, session.id, ackOffset)
								if err != nil {
									select {
									case <-ctx.Done():
										return nil
									default:
									}
									session.logger.Error("failed to ack message on queues service", zap.Error(err))
								}
							case 1:
								ingressQueue.Put(ctx, message, func() {
									err = local.queues.AckMessage(ctx, session.id, ackOffset)
									if err != nil {
										select {
										case <-ctx.Done():
											return
										default:
										}
										session.logger.Error("failed to ack message on queues service", zap.Error(err))
									}
								})
							default:
							}
							return nil
						})
						if err == nil {
							return nil
						}
						select {
						case <-ctx.Done():
							return nil
						default:
						}
						if count >= 5 {
							return backoff.Permanent(err)
						}
						count++
						return err
					}, backoff.NewExponentialBackOff())
					if err != nil {
						session.logger.Error("failed to poll queues for messages", zap.Error(err))
					}
					poller <- err
				}()
			}
			err = enc.SubAck(suback)
			if err != nil {
				cancelReqCtx()
				return err
			}
		case *packet.Unsubscribe:
			if session.token == "" {
				cancelReqCtx()
				return ErrConnectNotDone
			}
			unsuback, err := local.broker.Unsubscribe(packetCtx, session.token, p)
			if err != nil {
				cancelReqCtx()
				session.logger.Error("failed to unsubscribe", zap.Error(err))
				continue
			}
			err = enc.UnsubAck(unsuback)
			if err != nil {
				cancelReqCtx()
				return err
			}
		case *packet.PubAck:
			if session.token == "" {
				cancelReqCtx()
				return ErrConnectNotDone
			}
			err := ingressQueue.Ack(p)
			if err != nil {
				cancelReqCtx()
				return err
			}
		case *packet.PingReq:
			if session.token == "" {
				cancelReqCtx()
				return ErrConnectNotDone
			}
			pingresp, err := local.broker.PingReq(packetCtx, session.token, p)
			if err != nil {
				cancelReqCtx()
				return err
			}
			err = session.encoder.PingResp(pingresp)
			if err != nil {
				cancelReqCtx()
				return err
			}
		case *packet.Disconnect:
			if session.token == "" {
				cancelReqCtx()
				return ErrConnectNotDone
			}
			local.broker.Disconnect(packetCtx, session.token, p)
			cancelReqCtx()
			return nil
		}
		select {
		case err := <-poller:
			session.logger.Error("failed to poll messages", zap.Error(err))
			cancelReqCtx()
			return err
		default:
		}
		err := renewDeadline(timer, t.Channel)
		if err != nil {
			session.logger.Error("failed to set connection deadline", zap.Error(err))
		}
		cancelReqCtx()
	}
	return dec.Err()
}
