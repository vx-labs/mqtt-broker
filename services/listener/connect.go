package listener

import (
	"context"
	"time"

	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func (local *endpoint) ConnectHandler(ctx context.Context, session *localSession, p *packet.Connect) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	id, token, connack, err := local.broker.Connect(ctx, session.transport, p)
	if err != nil {
		session.logger.Error("CONNECT failed", zap.Error(err))
		session.encoder.ConnAck(connack)
		return ErrConnectNotDone
	}
	if connack.ReturnCode != packet.CONNACK_CONNECTION_ACCEPTED {
		session.encoder.ConnAck(connack)
		return ErrConnectNotDone
	}
	if id == "" {
		session.logger.Error("broker returned an empty session id")
		session.encoder.ConnAck(&packet.ConnAck{
			ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
			Header:     &packet.Header{},
		})
		return ErrConnectNotDone
	}
	if token == "" {
		session.logger.Error("broker returned an empty session token")
		session.encoder.ConnAck(&packet.ConnAck{
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
		old.(*localSession).transport.Channel.Close()
	}
	err = session.encoder.ConnAck(&packet.ConnAck{
		ReturnCode: packet.CONNACK_CONNECTION_ACCEPTED,
		Header:     &packet.Header{},
	})
	if err != nil {
		return err
	}
	session.timer = p.KeepaliveTimer
	session.logger.Info("started session")
	return nil
}
