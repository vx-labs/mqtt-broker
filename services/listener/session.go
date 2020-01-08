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
		transport: t,
		logger:    logger,
		cancel:    cancel,
	}
	defer func() {
		t.Channel.Close()
		local.mutex.Lock()
		local.sessions.Delete(session)
		local.mutex.Unlock()
		cancel()
	}()
	err := local.handleSessionPackets(ctx, session)
	if err != nil {
		if ctx.Err() == context.Canceled {
			session.logger.Info("session canceled")
			return
		}
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

func (local *endpoint) handleSessionPackets(ctx context.Context, session *localSession) error {
	session.logger.Info("accepted new connection")
	t := session.transport
	enc := encoder.New(t.Channel)
	session.encoder = enc
	session.timer = CONNECT_DEADLINE
	dec := decoder.Async(t.Channel)
	ingressQueue := inflight.New(enc.Publish)
	session.inflights = ingressQueue
	defer ingressQueue.Close()

	renewDeadline(session.timer, t.Channel)
	defer func() {
		dec.Cancel()
	}()

	if p, ok := (<-dec.Packet()).(*packet.Connect); ok {
		err := local.ConnectHandler(ctx, session, p)
		if err != nil {
			return err
		}
	} else {
		return ErrConnectNotDone
	}
	poller := make(chan error)
	go local.startPoller(ctx, session, poller)

	for data := range dec.Packet() {
		if sessionTracePacket == "true" {
			session.logger.Debug("session sent packet", zap.Any("traced_packet", data))
		}
		switch p := data.(type) {
		case *packet.Publish:
			err := local.PublishHandler(ctx, session, p)
			if err != nil {
				session.logger.Error("failed to publish", zap.Error(err))
			}
		case *packet.Subscribe:
			err := local.SubscribeHandler(ctx, session, p)
			if err != nil {
				session.logger.Error("failed to subscribe", zap.Error(err))
			}
		case *packet.Unsubscribe:
			err := local.UnsubscribeHandler(ctx, session, p)
			if err != nil {
				session.logger.Error("failed to unsubscribe", zap.Error(err))
			}
		case *packet.PubAck:
			err := ingressQueue.Ack(p)
			if err != nil {
				return err
			}
		case *packet.PingReq:
			pingresp, err := local.broker.PingReq(ctx, session.refreshToken, p)
			if err != nil {
				return err
			}
			err = session.encoder.PingResp(pingresp)
			if err != nil {
				return err
			}
		case *packet.Disconnect:
			local.broker.Disconnect(ctx, session.token, p)
			return nil
		}
		select {
		case err := <-poller:
			session.logger.Error("failed to poll messages", zap.Error(err))
			return err
		default:
		}
		err := renewDeadline(session.timer, t.Channel)
		if err != nil {
			session.logger.Error("failed to set connection deadline", zap.Error(err))
		}
	}
	return dec.Err()
}
