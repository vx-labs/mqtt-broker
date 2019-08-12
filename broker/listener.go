package broker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/vx-labs/mqtt-broker/cluster"
	listenerpb "github.com/vx-labs/mqtt-broker/listener/pb"
	publishQueue "github.com/vx-labs/mqtt-broker/queues/publish"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
	"github.com/vx-labs/mqtt-broker/topics"
	"github.com/vx-labs/mqtt-broker/transport"
	"github.com/vx-labs/mqtt-protocol/packet"
)

var enforceClientIDUniqueness = true

func validateClientID(clientID []byte) bool {
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

func (b *Broker) Connect(ctx context.Context, metadata transport.Metadata, p *packet.Connect) (string, string, *packet.ConnAck, error) {
	sessionID := newUUID()
	out := connectReturn{
		connack: &packet.ConnAck{
			Header:     p.Header,
			ReturnCode: packet.CONNACK_REFUSED_SERVER_UNAVAILABLE,
		},
	}
	done := make(chan struct{})
	err := b.workers.Call(func() error {
		defer close(done)
		clientID := p.ClientId
		clientIDstr := string(clientID)
		if !validateClientID(clientID) {
			out.connack.ReturnCode = packet.CONNACK_REFUSED_IDENTIFIER_REJECTED
			return fmt.Errorf("invalid client ID")
		}
		tenant, err := b.Authenticate(metadata, clientID, string(p.Username), string(p.Password))
		if err != nil {
			out.connack.ReturnCode = packet.CONNACK_REFUSED_BAD_USERNAME_OR_PASSWORD
			return fmt.Errorf("WARN: authentication failed for client ID %q: %v", p.ClientId, err)
		}
		if enforceClientIDUniqueness {
			set, err := b.Sessions.ByClientID(clientIDstr)
			if err != nil {
				return fmt.Errorf("WARN: authentication failed for client ID %q: %v", clientIDstr, err)
			}
			if len(set) > 0 {
				if err := set.ApplyE(func(session sessions.Session) error {
					b.Sessions.Delete(session.ID, "session_disconnected")
					if b.isSessionLocal(session) {
						b.logger.Info("closing old session to free client-id", zap.String("session_id", session.ID), zap.String("client_id", session.ClientID))
						session.Transport.Close()
					}
					return nil
				}); err != nil {
					out.connack.ReturnCode = packet.CONNACK_REFUSED_IDENTIFIER_REJECTED
					return fmt.Errorf("invalid client ID")
				}
			}
		}
		sess := sessions.Session{
			Metadata: sessions.Metadata{
				ID:                sessionID,
				ClientID:          clientIDstr,
				Created:           time.Now().Unix(),
				Tenant:            tenant,
				Peer:              metadata.Endpoint,
				WillPayload:       p.WillPayload,
				WillQoS:           p.WillQos,
				WillRetain:        p.WillRetain,
				WillTopic:         p.WillTopic,
				Transport:         metadata.Name,
				RemoteAddress:     metadata.RemoteAddress,
				KeepaliveInterval: p.KeepaliveTimer,
			},
		}
		err = b.Sessions.Upsert(sess, b.RemoteRPCProvider(sessionID, sess.Peer))
		if err != nil {
			return err
		}
		token, err := EncodeSessionToken(os.Getenv("JWT_SIGN_KEY"), sess)
		if err != nil {
			return err
		}
		out.token = token
		b.logger.Info("session connected", zap.String("session_id", sessionID), zap.String("client_id", string(p.ClientId)), zap.String("username", string(p.Username)), zap.String("remote_address", metadata.RemoteAddress), zap.String("transport", metadata.Name))

		out.connack.ReturnCode = packet.CONNACK_CONNECTION_ACCEPTED
		return nil
	})
	<-done
	return sessionID, out.token, out.connack, err
}
func (b *Broker) Subscribe(ctx context.Context, id string, p *packet.Subscribe) (*packet.SubAck, error) {
	sess, err := b.Sessions.ByID(id)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", id), zap.String("packet", "subscribe"))
		return nil, err
	}
	for idx, pattern := range p.Topic {
		subID := makeSubID(id, pattern)
		event := subscriptions.Subscription{
			Metadata: subscriptions.Metadata{
				ID:        subID,
				Pattern:   pattern,
				Qos:       p.Qos[idx],
				Tenant:    sess.Tenant,
				SessionID: sess.ID,
				Peer:      sess.Peer,
			},
		}
		err := b.Subscriptions.Create(event, func(ctx context.Context, publish packet.Publish) error {
			if (publish.Header.Qos) > event.Qos {
				publish.Header.Qos = event.Qos
			}
			return sess.Transport.Publish(ctx, &publish)
		})
		if err != nil {
			return nil, err
		}
		b.logger.Info("session subscribed",
			zap.String("session_id", sess.ID),
			zap.String("subscription_id", subID),
			zap.Int32("qos", p.Qos[idx]),
			zap.Binary("topic_pattern", event.Pattern))

		// Look for retained messages
		set, err := b.Topics.ByTopicPattern(sess.Tenant, pattern)
		if err != nil {
			return nil, err
		}
		packetQoS := p.Qos[idx]
		go func() {
			set.Apply(func(message topics.RetainedMessage) {
				qos := getLowerQoS(message.Qos, packetQoS)
				sess.Transport.Publish(b.ctx, &packet.Publish{
					Header: &packet.Header{
						Qos:    qos,
						Retain: true,
					},
					MessageId: 1,
					Payload:   message.Payload,
					Topic:     message.Topic,
				})
			})
		}()
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

func (b *Broker) routeMessage(tenant string, p *packet.Publish) error {
	recipients, err := b.Subscriptions.ByTopic(tenant, p.Topic)
	if err != nil {
		return err
	}
	message := *p
	message.Header.Retain = false
	if len(recipients) > 0 {
		recipients.Apply(func(sub subscriptions.Subscription) {
			sub.Sender(b.ctx, message)
		})
	}
	return nil
}
func (b *Broker) Publish(ctx context.Context, id string, p *packet.Publish) (*packet.PubAck, error) {
	session, err := b.Sessions.ByID(id)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", id), zap.String("packet", "publish"))
		return nil, err
	}
	if p.Header.Qos == 2 {
		err = errors.New("QoS2 is not supported")
		return nil, err
	}
	b.publishQueue.Enqueue(&publishQueue.Message{
		Tenant:  session.Tenant,
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
func (b *Broker) Unsubscribe(ctx context.Context, id string, p *packet.Unsubscribe) (*packet.UnsubAck, error) {
	sess, err := b.Sessions.ByID(id)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", id), zap.String("packet", "unsubscribe"))
		return nil, err
	}
	set, err := b.Subscriptions.BySession(sess.ID)
	if err != nil {
		return nil, err
	}
	set = set.Filter(func(sub subscriptions.Subscription) bool {
		for _, topic := range p.Topic {
			if bytes.Compare(topic, sub.Pattern) == 0 {
				return true
			}
		}
		return false
	})
	set.Apply(func(sub subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})
	return &packet.UnsubAck{
		MessageId: p.MessageId,
		Header:    &packet.Header{},
	}, nil
}
func (b *Broker) Disconnect(ctx context.Context, id string, p *packet.Disconnect) error {
	sess, err := b.Sessions.ByID(id)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", id), zap.String("packet", "disconnect"))
		return err
	}
	b.Sessions.Delete(id, "session_disconnected")
	err = b.deleteSessionSubscriptions(sess)
	if err != nil {
		b.logger.Error("failed to delete session subscriptions when disconnecting", zap.String("session_id", id), zap.Error(err))
	}
	b.logger.Info("session disconnected", zap.String("session_id", sess.ID))
	return nil
}

func (b *Broker) CloseSession(ctx context.Context, id string) error {
	sess, err := b.Sessions.ByID(id)
	if err != nil {
		if err == sessions.ErrSessionNotFound {
			return nil
		}
		return err
	}
	if len(sess.WillTopic) > 0 {
		if sess.WillRetain {
			retainedMessage := topics.RetainedMessage{
				Metadata: topics.Metadata{
					Payload: sess.WillPayload,
					Qos:     sess.WillQoS,
					Tenant:  sess.Tenant,
					Topic:   sess.WillTopic,
				},
			}
			b.Topics.Create(retainedMessage)
			if err != nil {
				b.logger.Warn("failed to retain LWT", zap.String("session_id", sess.ID), zap.Error(err))
			}
		}
		b.routeMessage(sess.Tenant, &packet.Publish{
			Header: &packet.Header{
				Dup:    false,
				Retain: false,
				Qos:    sess.WillQoS,
			},
			Payload: sess.WillPayload,
			Topic:   sess.WillTopic,
		})
	}
	b.Sessions.Delete(id, "session_lost")
	err = b.deleteSessionSubscriptions(sess)
	if err != nil {
		b.logger.Error("failed to delete session subscriptions", zap.String("session_id", sess.ID), zap.Error(err))
		return err
	}
	b.logger.Info("session lost", zap.String("session_id", sess.ID))
	return nil
}

func (b *Broker) PingReq(ctx context.Context, id string, _ *packet.PingReq) (*packet.PingResp, error) {
	_, err := b.Sessions.ByID(id)
	if err != nil {
		b.logger.Warn("received packet from an unknown session", zap.String("session_id", id), zap.String("packet", "pingreq"))
		return nil, err
	}
	return &packet.PingResp{
		Header: &packet.Header{},
	}, nil
}
