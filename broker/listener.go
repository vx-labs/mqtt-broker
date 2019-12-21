package broker

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/vx-labs/mqtt-broker/events"

	"github.com/vx-labs/mqtt-broker/cluster"
	"go.uber.org/zap"

	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
	"github.com/vx-labs/mqtt-broker/transport"
	"github.com/vx-labs/mqtt-protocol/packet"
)

var enforceClientIDUniqueness = true

var (
	ErrAuthenticationFailed = errors.New("authentication failed")
)

func isClientIDValid(clientID []byte) bool {
	return len(clientID) > 0 && len(clientID) < 128
}

type localTransport struct {
	ctx  context.Context
	id   string
	mesh cluster.Mesh
	peer string
}

type connectReturn struct {
	token   string
	connack *packet.ConnAck
}

func (b *Broker) SigningKey() string {
	return os.Getenv("JWT_SIGN_KEY")
}
func connack(code int32) *packet.ConnAck {
	return &packet.ConnAck{
		Header:     &packet.Header{},
		ReturnCode: code,
	}
}

func (b *Broker) Connect(ctx context.Context, metadata transport.Metadata, p *packet.Connect) (string, string, *packet.ConnAck, error) {
	sessionID := newUUID()
	clientID := p.ClientId
	clientIDstr := string(clientID)
	logger := b.logger.With(zap.String("session_id", sessionID), zap.String("client_id", string(clientIDstr)), zap.String("username", string(p.Username)), zap.String("remote_address", metadata.RemoteAddress), zap.String("transport", metadata.Name))
	if !isClientIDValid(clientID) {
		logger.Info("connection refused: invalid client id")
		return "", "", connack(packet.CONNACK_REFUSED_IDENTIFIER_REJECTED), nil
	}
	tenant, err := b.Authenticate(metadata, clientID, string(p.Username), string(p.Password))
	if err != nil {
		logger.Info("authentication failed", zap.Error(err))
		return "", "", connack(packet.CONNACK_REFUSED_BAD_USERNAME_OR_PASSWORD), nil
	}
	input := sessions.SessionCreateInput{
		ID:                sessionID,
		ClientID:          clientIDstr,
		Tenant:            tenant,
		Peer:              metadata.Endpoint,
		WillPayload:       p.WillPayload,
		WillQoS:           p.WillQos,
		WillRetain:        p.WillRetain,
		WillTopic:         p.WillTopic,
		Transport:         metadata.Name,
		RemoteAddress:     metadata.RemoteAddress,
		KeepaliveInterval: p.KeepaliveTimer,
		Timestamp:         time.Now().Unix(),
	}
	events.Commit(ctx, b.Messages, sessionID, &events.StateTransition{
		Event: &events.StateTransition_SessionCreated{
			SessionCreated: &events.SessionCreated{
				ID:                input.ID,
				Tenant:            input.Tenant,
				ClientID:          input.ClientID,
				KeepaliveInterval: input.KeepaliveInterval,
				Peer:              metadata.Endpoint,
				WillPayload:       p.WillPayload,
				WillQoS:           p.WillQos,
				WillRetain:        p.WillRetain,
				WillTopic:         p.WillTopic,
				Transport:         metadata.Name,
				RemoteAddress:     metadata.RemoteAddress,
				Timestamp:         time.Now().Unix(),
			},
		},
	})

	err = b.Queues.Create(b.ctx, sessionID)
	if err != nil {
		logger.Error("failed to create queue", zap.Error(err))
		return "", "", nil, err
	}
	token, err := EncodeSessionToken(b.SigningKey(), input.Tenant, input.ID)
	if err != nil {
		logger.Error("failed to encode session JWT", zap.Error(err))
		return "", "", nil, err
	}
	logger.Info("session connected")
	return sessionID, token, connack(packet.CONNACK_CONNECTION_ACCEPTED), nil
}
func (b *Broker) Subscribe(ctx context.Context, token string, p *packet.Subscribe) (*packet.SubAck, error) {
	sess, err := DecodeSessionToken(b.SigningKey(), token)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", sess.SessionID), zap.String("packet", "subscribe"))
		return nil, err
	}
	transition := []*events.StateTransition{}
	for idx, pattern := range p.Topic {
		subID := makeSubID(sess.SessionID, pattern)
		transition = append(transition, &events.StateTransition{
			Event: &events.StateTransition_SessionSubscribed{
				SessionSubscribed: &events.SessionSubscribed{
					SessionID: sess.SessionID,
					Qos:       p.Qos[idx],
					Tenant:    sess.SessionTenant,
					Pattern:   pattern,
				},
			},
		})
		b.logger.Info("session subscribed",
			zap.String("session_id", sess.SessionID),
			zap.String("subscription_id", subID),
			zap.Int32("qos", p.Qos[idx]),
			zap.Binary("topic_pattern", pattern))

		// go func(packetQoS int32, pattern []byte) {
		// 	set, err := b.Topics.ByTopicPattern(b.ctx, sess.SessionTenant, pattern)
		// 	if err != nil {
		// 		b.logger.Error("failed to look for retained messages",
		// 			zap.Error(err),
		// 			zap.String("session_id", sess.SessionID))
		// 		return
		// 	}
		// 	for _, message := range set {
		// 		qos := getLowerQoS(message.Qos, packetQoS)
		// 		err := b.sendToSession(b.ctx, sess.SessionID, &packet.Publish{
		// 			Header: &packet.Header{
		// 				Qos:    qos,
		// 				Retain: true,
		// 			},
		// 			MessageId: 1,
		// 			Payload:   message.Payload,
		// 			Topic:     message.Topic,
		// 		})
		// 		if err != nil {
		// 			b.logger.Error("failed to publish retained message",
		// 				zap.Error(err),
		// 				zap.String("session_id", sess.SessionID),
		// 				zap.String("subscription_id", subID),
		// 				zap.Binary("topic_pattern", message.Topic))
		// 		}
		// 	}
		// }(p.Qos[idx], pattern)
	}
	err = events.Commit(ctx, b.Messages, sess.SessionID, transition...)
	if err != nil {
		b.logger.Error("failed to commit session subscription event", zap.Error(err))
		return nil, err
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
	return &packet.SubAck{
		Header:    &packet.Header{},
		MessageId: p.MessageId,
		Qos:       qos,
	}, nil
}

func (b *Broker) Publish(ctx context.Context, token string, p *packet.Publish) (*packet.PubAck, error) {
	sess, err := DecodeSessionToken(b.SigningKey(), token)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", sess.SessionID), zap.String("packet", "publish"))
		return nil, err
	}
	if p.Header.Qos == 2 {
		err = errors.New("QoS2 is not supported")
		return nil, err
	}
	err = b.enqueuePublish(sess.SessionTenant, sess.SessionID, p)
	if err != nil {
		return nil, err
	}
	if p.Header.Qos == 1 {
		puback := &packet.PubAck{
			Header:    &packet.Header{},
			MessageId: p.MessageId,
		}
		return puback, nil
	}
	return nil, nil

}
func (b *Broker) Unsubscribe(ctx context.Context, token string, p *packet.Unsubscribe) (*packet.UnsubAck, error) {
	sess, err := DecodeSessionToken(b.SigningKey(), token)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", sess.SessionID), zap.String("packet", "unsubscribe"))
		return nil, err
	}
	transition := []*events.StateTransition{}
	for _, pattern := range p.Topic {
		transition = append(transition, &events.StateTransition{
			Event: &events.StateTransition_SessionUnsubscribed{
				SessionUnsubscribed: &events.SessionUnsubscribed{
					SessionID: sess.SessionID,
					Tenant:    sess.SessionTenant,
					Pattern:   pattern,
				},
			},
		})
	}
	err = events.Commit(ctx, b.Messages, sess.SessionID, transition...)
	if err != nil {
		b.logger.Error("failed to commit session subscription event", zap.Error(err))
		return nil, err
	}
	return &packet.UnsubAck{
		MessageId: p.MessageId,
		Header:    &packet.Header{},
	}, nil
}
func (b *Broker) Disconnect(ctx context.Context, token string, p *packet.Disconnect) error {
	sess, err := DecodeSessionToken(b.SigningKey(), token)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", sess.SessionID), zap.String("packet", "disconnect"))
		return err
	}
	return events.Commit(ctx, b.Messages, sess.SessionID, &events.StateTransition{
		Event: &events.StateTransition_SessionClosed{
			SessionClosed: &events.SessionClosed{
				ID:     sess.SessionID,
				Tenant: sess.SessionTenant,
			},
		},
	})
}

func (b *Broker) CloseSession(ctx context.Context, token string) error {
	decodedToken, err := DecodeSessionToken(b.SigningKey(), token)
	if err != nil {
		return err
	}
	return events.Commit(ctx, b.Messages, decodedToken.SessionID, &events.StateTransition{
		Event: &events.StateTransition_SessionLost{
			SessionLost: &events.SessionLost{
				ID:     decodedToken.SessionID,
				Tenant: decodedToken.SessionTenant,
			},
		},
	})
}

func (b *Broker) PingReq(ctx context.Context, id string, _ *packet.PingReq) (*packet.PingResp, error) {
	token, err := DecodeSessionToken(b.SigningKey(), id)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", token.SessionID), zap.String("packet", "pingreq"))
		return nil, err
	}
	err = b.Sessions.RefreshKeepAlive(ctx, token.SessionID, time.Now().Unix())
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", token.SessionID), zap.String("packet", "pingreq"))
		return nil, err
	}

	return &packet.PingResp{
		Header: &packet.Header{},
	}, nil
}
