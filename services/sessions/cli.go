package sessions

import (
	"context"
	fmt "fmt"
	"net"
	"time"

	proto "github.com/golang/protobuf/proto"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/vx-labs/mqtt-broker/adapters/ap"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-broker/network"
	broker "github.com/vx-labs/mqtt-broker/services/broker/pb"
	kv "github.com/vx-labs/mqtt-broker/services/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/services/sessions/pb"
	"github.com/vx-labs/mqtt-broker/stream"
	"github.com/vx-labs/mqtt-protocol/packet"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	b.stream.Shutdown()
	b.state.Shutdown()
	b.gprcServer.GracefulStop()
}

func contains(needle string, haystack []string) bool {
	for idx := range haystack {
		if needle == haystack[idx] {
			return true
		}
	}
	return false
}

func (b *server) Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error {
	b.store = NewSessionStore(logger)
	service := catalog.Service(fmt.Sprintf("%sgossip", name))
	userService := catalog.Service(name)
	b.state = ap.GossipDistributer(id, service, b.store, logger)
	listener, err := userService.ListenTCP()
	if err != nil {
		return err
	}
	b.listener = listener
	kvConn, err := catalog.Dial("kv")
	if err != nil {
		return err
	}
	messagesConn, err := catalog.Dial("messages")
	if err != nil {
		return err
	}
	k := kv.NewClient(kvConn)
	m := messages.NewClient(messagesConn)
	b.stream = stream.NewClient(k, m, logger)
	b.Messages = m
	ctx := context.Background()
	expirationTicker := time.NewTicker(10 * time.Second)
	go func() {
		lockPath := []byte("sessions/expiration_lock")
		for range expirationTicker.C {
			now := time.Now().Unix()
			sessions, err := b.store.All(&pb.SessionFilterInput{})
			if err != nil {
				logger.Error("failed to list sessions", zap.Error(err))
				continue
			}
			sessionExpiredEvents := []*events.StateTransition{}
			for _, e := range sessions.Sessions {
				if isSessionExpired(e, now) {
					sessionExpiredEvents = append(sessionExpiredEvents, &events.StateTransition{
						Event: &events.StateTransition_SessionLost{
							SessionLost: &events.SessionLost{
								ID:     e.ID,
								Tenant: e.Tenant,
							},
						},
					})
				}
			}
			if len(sessionExpiredEvents) > 0 {
				payload, err := events.Encode(sessionExpiredEvents...)
				if err != nil {
					logger.Error("failed to encode session expired events", zap.Error(err))
					continue
				}
				v, md, err := k.GetWithMetadata(ctx, lockPath)
				if err != nil {
					logger.Error("failed to read session expiration lock key", zap.Error(err))
					continue
				}
				if v != nil {
					continue
				}
				err = k.SetWithVersion(ctx, lockPath, []byte(b.id), md.Version, kv.WithTimeToLive(5*time.Second))
				if err != nil {
					logger.Error("failed to create session expiration lock key", zap.Error(err))
					continue
				}
				err = m.Put(ctx, "events", b.id, payload)
				if err != nil {
					logger.Error("failed to enqueue event in message store", zap.Error(err))
					continue
				}
				k.DeleteWithVersion(ctx, lockPath, md.Version+1)
				logger.Info("expired sessions", zap.Int("expired_session_count", len(sessionExpiredEvents)))
			}
		}
	}()
	b.stream.ConsumeStream(ctx, "events", b.consumeStream,
		stream.WithConsumerID(b.id),
		stream.WithConsumerGroupID("sessions"),
		stream.WithInitialOffsetBehaviour(stream.OFFSET_BEHAVIOUR_FROM_START),
	)
	return nil
}

func (b *server) consumeStream(messages []*messages.StoredMessage) (int, error) {
	if b.store == nil {
		return 0, errors.New("store not ready")
	}
	for idx := range messages {
		eventSet, err := events.Decode(messages[idx].Payload)
		if err != nil {
			return idx, errors.Wrap(err, "failed to decode message for shard")
		}
		for _, eventPayload := range eventSet {
			switch event := eventPayload.GetEvent().(type) {
			case *events.StateTransition_SessionClosed:
				input := event.SessionClosed
				err = b.store.Delete(input.ID)
				if err != nil {
					b.logger.Warn("failed to delete session", zap.Error(err))
				}
			case *events.StateTransition_SessionKeepalived:
				input := event.SessionKeepalived
				err := b.store.Update(input.SessionID, func(session pb.Session) *pb.Session {
					session.LastKeepAlive = input.Timestamp
					return &session
				})
				if err != nil {
					b.logger.Warn("failed to update session last keepalived", zap.String("session_id", input.SessionID), zap.Error(err))
				}
			case *events.StateTransition_SessionLost:
				input := event.SessionLost
				oldSession, err := b.store.ByID(input.ID)
				if err != nil {
					continue
				}
				err = b.store.Delete(input.ID)
				if err != nil {
					b.logger.Warn("failed to delete session", zap.Error(err))
				}
				err = b.maybeSendWill(oldSession)
				if err != nil {
					b.logger.Error("failed to enqueue LWT message in message store", zap.Error(err))
					return idx, err
				}
			case *events.StateTransition_SessionCreated:
				input := event.SessionCreated

				oldSessions, err := b.store.ByClientID(input.ClientID)
				if err == nil {
					sessionReplacedEvents := []*events.StateTransition{}
					for _, e := range oldSessions.Sessions {
						sessionReplacedEvents = append(sessionReplacedEvents, &events.StateTransition{
							Event: &events.StateTransition_SessionClosed{
								SessionClosed: &events.SessionClosed{
									ID:     e.ID,
									Tenant: e.Tenant,
								},
							},
						})
					}
					payload, err := events.Encode(sessionReplacedEvents...)
					if err != nil {
						b.logger.Error("failed to encode session replaced events", zap.Error(err))
					} else {
						err = b.Messages.Put(b.ctx, "events", b.id, payload)
						if err != nil {
							b.logger.Error("failed to enqueue event in message store", zap.Error(err))
						}
					}
				}
				err = b.store.Create(&pb.Session{
					ClientID:          input.ClientID,
					ID:                input.ID,
					KeepaliveInterval: input.KeepaliveInterval,
					Peer:              input.Peer,
					RemoteAddress:     input.RemoteAddress,
					Tenant:            input.Tenant,
					Transport:         input.Transport,
					WillPayload:       input.WillPayload,
					WillTopic:         input.WillTopic,
					WillRetain:        input.WillRetain,
					WillQoS:           input.WillQoS,
					Created:           input.Timestamp,
					LastKeepAlive:     input.Timestamp,
				})
				if err != nil {
					b.logger.Error("failed to create session", zap.Error(err))
					return idx, err
				}
			}
		}

	}
	return len(messages), nil
}

func (b *server) maybeSendWill(oldSession *pb.Session) error {
	if len(oldSession.WillTopic) > 0 {
		payload := &broker.MessagePublished{
			Tenant: oldSession.Tenant,
			Publish: &packet.Publish{
				Header: &packet.Header{
					Retain: oldSession.WillRetain,
					Qos:    oldSession.WillQoS,
				},
				Topic:   oldSession.WillTopic,
				Payload: oldSession.WillPayload,
			},
		}
		data, err := proto.Marshal(payload)
		if err != nil {
			return err
		}
		err = b.Messages.Put(b.ctx, "messages", oldSession.ID, data)
		if err != nil {
			b.logger.Error("failed to enqueue LWT message in message store", zap.Error(err))
			return err
		}
	}
	return nil
}
func (m *server) Health() string {
	if m.state == nil {
		return "warning"
	}
	return m.state.Health()
}
func (m *server) Serve(port int) net.Listener {
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterSessionsServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(m.listener)
	m.gprcServer = s
	return m.listener
}
