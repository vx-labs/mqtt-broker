package broker

import (
	"bytes"
	"context"
	"errors"
	"os"
	"time"

	"github.com/vx-labs/mqtt-broker/cluster"
	listenerpb "github.com/vx-labs/mqtt-broker/listener/pb"
	publishQueue "github.com/vx-labs/mqtt-broker/queues/publish"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
	subscriptions "github.com/vx-labs/mqtt-broker/subscriptions/pb"
	"github.com/vx-labs/mqtt-broker/topics"
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

func (local *localTransport) Close() error {
	return local.mesh.DialAddress("listener", local.peer, func(conn *grpc.ClientConn) error {
		c := listenerpb.NewClient(conn)
		return c.Shutdown(local.ctx, local.id)
	})
}

func (local *localTransport) Publish(ctx context.Context, p *packet.Publish) error {
	return local.mesh.DialAddress("listener", local.peer, func(conn *grpc.ClientConn) error {
		c := listenerpb.NewClient(conn)
		return c.SendPublish(ctx, local.id, p)
	})
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
	if enforceClientIDUniqueness {
		set, err := b.Sessions.ByClientID(b.ctx, clientIDstr)
		if err != nil {
			logger.Error("failed to lookup for other sessions using same client id", zap.Error(err))
			return "", "", nil, err
		}
		if len(set) > 0 {
			for _, session := range set {
				b.Sessions.Delete(b.ctx, session.ID)
			}
		}
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
	err = b.Sessions.Create(b.ctx, input)
	if err != nil {
		logger.Error("failed to create session", zap.Error(err))
		return "", "", nil, err
	}
	sess, err := b.Sessions.ByID(b.ctx, input.ID)
	if err != nil {
		logger.Error("failed to read session", zap.Error(err))
		return "", "", nil, err
	}
	token, err := EncodeSessionToken(b.SigningKey(), sess)
	if err != nil {
		logger.Error("failed to encode session JWT", zap.Error(err))
		return "", "", nil, err
	}
	logger.Info("session connected")
	return sessionID, token, connack(packet.CONNACK_CONNECTION_ACCEPTED), nil
}
func (b *Broker) sendToSession(ctx context.Context, id string, peer string, p *packet.Publish) error {
	return b.mesh.DialAddress("listener", peer, func(conn *grpc.ClientConn) error {
		c := listenerpb.NewClient(conn)
		return c.SendPublish(ctx, id, p)
	})
}

func (b *Broker) Subscribe(ctx context.Context, token string, p *packet.Subscribe) (*packet.SubAck, error) {
	sess, err := DecodeSessionToken(b.SigningKey(), token)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", sess.SessionID), zap.String("packet", "subscribe"))
		return nil, err
	}
	for idx, pattern := range p.Topic {
		subID := makeSubID(sess.SessionID, pattern)
		event := subscriptions.SubscriptionCreateInput{
			ID:        subID,
			Pattern:   pattern,
			Qos:       p.Qos[idx],
			Tenant:    sess.SessionTenant,
			SessionID: sess.SessionID,
			Peer:      sess.PeerID,
		}
		err := b.Subscriptions.Create(b.ctx, event)
		if err != nil {
			return nil, err
		}
		b.logger.Info("session subscribed",
			zap.String("session_id", sess.SessionID),
			zap.String("subscription_id", subID),
			zap.Int32("qos", p.Qos[idx]),
			zap.Binary("topic_pattern", event.Pattern))

		// Look for retained messages
		set, err := b.Topics.ByTopicPattern(sess.SessionTenant, pattern)
		if err != nil {
			return nil, err
		}
		packetQoS := p.Qos[idx]
		set.Apply(func(message topics.RetainedMessage) {
			qos := getLowerQoS(message.Qos, packetQoS)
			err := b.sendToSession(b.ctx, sess.SessionID, sess.PeerID, &packet.Publish{
				Header: &packet.Header{
					Qos:    qos,
					Retain: true,
				},
				MessageId: 1,
				Payload:   message.Payload,
				Topic:     message.Topic,
			})
			if err != nil {
				b.logger.Error("failed to publish retained message",
					zap.Error(err),
					zap.String("session_id", sess.SessionID),
					zap.String("subscription_id", subID),
					zap.Binary("topic_pattern", message.Topic))
			}
		})
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
	b.publishQueue.Enqueue(&publishQueue.Message{
		Tenant:  sess.SessionTenant,
		Publish: p,
	})
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
	userSubscriptions, err := b.Subscriptions.BySession(b.ctx, sess.SessionID)
	if err != nil {
		return nil, err
	}
	for _, subscription := range userSubscriptions {
		for _, topic := range p.Topic {
			if bytes.Compare(topic, subscription.Pattern) == 0 {
				b.Subscriptions.Delete(b.ctx, subscription.ID)
			}
		}
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
	err = b.Sessions.Delete(b.ctx, sess.SessionID)
	if err != nil {
		b.logger.Error("failed to delete session when disconnecting", zap.String("session_id", sess.SessionID), zap.Error(err))
	}
	b.logger.Info("session disconnected", zap.String("session_id", sess.SessionID))
	return nil
}

func (b *Broker) CloseSession(ctx context.Context, token string) error {
	decodedToken, err := DecodeSessionToken(b.SigningKey(), token)
	if err != nil {
		return err
	}
	sess, err := b.Sessions.ByID(b.ctx, decodedToken.SessionID)
	if err != nil {
		return nil
	}
	if len(sess.WillTopic) > 0 {
		b.logger.Info("sending LWT", zap.String("session_id", sess.ID), zap.Error(err))
		b.publishQueue.Enqueue(&publishQueue.Message{
			Tenant: sess.Tenant,
			Publish: &packet.Publish{
				Header: &packet.Header{
					Dup:    false,
					Retain: sess.WillRetain,
					Qos:    sess.WillQoS,
				},
				Payload: sess.WillPayload,
				Topic:   sess.WillTopic,
			},
		})
	}
	b.Sessions.Delete(b.ctx, decodedToken.SessionID)
	b.logger.Info("session lost", zap.String("session_id", decodedToken.SessionID))
	return nil
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
