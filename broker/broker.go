package broker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vx-labs/mqtt-broker/broker/cluster"

	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-broker/peers"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/topics"

	"github.com/vx-labs/mqtt-broker/broker/listener/transport"
	"github.com/vx-labs/mqtt-broker/broker/rpc"

	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/vx-labs/mqtt-broker/broker/listener"

	"github.com/vx-labs/mqtt-broker/identity"

	"github.com/vx-labs/mqtt-broker/subscriptions"
)

const (
	EVENT_MESSAGE_PUBLISHED = "message_published"
	EVENT_STATE_UPDATED     = "state_updated"
)

type PeerStore interface {
	ByID(id string) (peers.Peer, error)
	All() (peers.SubscriptionSet, error)
	Upsert(sess peers.Peer) error
	Delete(id string) error
	On(event string, handler func(peers.Peer)) func()
}
type SessionStore interface {
	ByID(id string) (sessions.Session, error)
	ByClientID(id string) (sessions.SessionSet, error)
	ByPeer(peer string) (sessions.SessionSet, error)
	All() (sessions.SessionSet, error)
	Exists(id string) bool
	Upsert(sess sessions.Session, transport sessions.Transport) error
	Delete(id, reason string) error
	On(event string, handler func(sessions.Session)) func()
}

type TopicStore interface {
	Create(message topics.RetainedMessage) error
	ByTopicPattern(tenant string, pattern []byte) (topics.RetainedMessageSet, error)
	All() (topics.RetainedMessageSet, error)
	On(event string, handler func(topics.RetainedMessage)) func()
}
type SubscriptionStore interface {
	ByTopic(tenant string, pattern []byte) (subscriptions.SubscriptionSet, error)
	ByID(id string) (subscriptions.Subscription, error)
	All() (subscriptions.SubscriptionSet, error)
	ByPeer(peer string) (subscriptions.SubscriptionSet, error)
	BySession(id string) (subscriptions.SubscriptionSet, error)
	Sessions() ([]string, error)
	Create(message subscriptions.Subscription, sender func(packet.Publish) error) error
	Delete(id string) error
	On(event string, handler func(subscriptions.Subscription)) func()
}
type Broker struct {
	ID            string
	authHelper    func(transport listener.Transport, sessionID []byte, username string, password string) (tenant string, err error)
	mesh          cluster.Mesh
	Subscriptions SubscriptionStore
	Sessions      SessionStore
	Topics        TopicStore
	Peers         PeerStore
	events        *events.Bus
	STANOutput    chan STANMessage
	Listener      io.Closer
	TCPTransport  io.Closer
	TLSTransport  io.Closer
	WSSTransport  io.Closer
	WSTransport   io.Closer
	RPC           net.Listener
	RPCCaller     *rpc.Caller
	workers       *Pool
}
type RemoteRPCTransport struct {
	peer string
	id   string
	rpc  *rpc.Caller
}

func (r *RemoteRPCTransport) Close() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return r.rpc.Call(r.peer, func(c rpc.BrokerServiceClient) error {
		_, err := c.CloseSession(ctx, &rpc.CloseSessionInput{
			ID: r.id,
		})
		return err
	})
}
func (r *RemoteRPCTransport) Publish(publish *packet.Publish) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return r.rpc.Call(r.peer, func(c rpc.BrokerServiceClient) error {
		_, err := c.DistributeMessage(ctx, &rpc.MessagePublished{
			Recipient: r.id,
			Dup:       publish.Header.Dup,
			Payload:   publish.Payload,
			Qos:       publish.Header.Qos,
			Retained:  publish.Header.Retain,
			Topic:     publish.Topic,
		})
		return err
	})
}

func (b *Broker) RemoteRPCProvider(addr, session string) sessions.Transport {
	return &RemoteRPCTransport{
		id:   session,
		peer: addr,
		rpc:  b.RPCCaller,
	}
}

func New(id identity.Identity, config Config) *Broker {
	broker := &Broker{
		ID:         id.ID(),
		authHelper: config.AuthHelper,
		events:     events.NewEventBus(),
		RPCCaller:  rpc.NewCaller(),
		workers:    NewPool(25),
	}

	l, listenerCh := listener.New(broker, config.Session.MaxInflightSize)
	broker.RPC = rpc.New(config.RPCPort, broker)
	if config.RPCIdentity == nil {
		_, port, err := net.SplitHostPort(broker.RPC.Addr().String())
		if err != nil {
			panic(err)
		}
		broker.mesh = cluster.MemberlistMesh(id, broker, cluster.NodeMeta{
			ID:      broker.ID,
			RPCAddr: fmt.Sprintf("%s:%s", id.Private().Host(), port),
		})
	} else {
		broker.mesh = cluster.MemberlistMesh(id, broker, cluster.NodeMeta{
			ID:      broker.ID,
			RPCAddr: config.RPCIdentity.Public().String(),
		})
	}
	hostedServices := []string{}
	hostedServices = append(hostedServices, "rpc-listener")
	subscriptionsStore, err := subscriptions.NewMemDBStore(broker.mesh, func(host string, session string, publish packet.Publish) error {
		ctx := context.Background()
		addr, err := broker.mesh.MemberRPCAddress(host)
		if err != nil {
			log.Printf("ERROR: failed to resove peer %s addr: %v", host, err)
			return err
		}
		return broker.RPCCaller.Call(addr, func(c rpc.BrokerServiceClient) error {
			_, err := c.DistributeMessage(ctx, &rpc.MessagePublished{
				Dup:       publish.Header.Dup,
				Payload:   publish.Payload,
				Qos:       publish.Header.Qos,
				Recipient: session,
				Topic:     publish.Topic,
			})
			if err != nil {
				log.Printf("ERROR: failed to send message to peer %s: %v", host, err)
			}
			return err
		})
	})
	if err != nil {
		log.Fatal(err)
	}
	hostedServices = append(hostedServices, "subscriptions-store")
	peersStore, err := peers.NewPeerStore(broker.mesh)
	if err != nil {
		log.Fatal(err)
	}
	hostedServices = append(hostedServices, "peers-store")
	topicssStore, err := topics.NewMemDBStore(broker.mesh)
	if err != nil {
		log.Fatal(err)
	}
	hostedServices = append(hostedServices, "topics-store")
	sessionsStore, err := sessions.NewSessionStore(broker.mesh, broker.RemoteRPCProvider, logrus.New().WithField("source", "session_store"))
	if err != nil {
		log.Fatal(err)
	}
	hostedServices = append(hostedServices, "sessions-store")
	broker.Peers = peersStore
	broker.Topics = topicssStore
	broker.Subscriptions = subscriptionsStore
	broker.Sessions = sessionsStore

	if config.TCPPort > 0 {
		tcpTransport, err := transport.NewTCPTransport(config.TCPPort, listenerCh)
		broker.TCPTransport = tcpTransport
		if err != nil {
			log.Printf("WARN: failed to start TCP listener on port %d: %v", config.TCPPort, err)
		} else {
			log.Printf("INFO: started TCP listener on port %d", config.TCPPort)
			hostedServices = append(hostedServices, "tcp-listener")
		}
	}
	if config.WSPort > 0 {
		wsTransport, err := transport.NewWSTransport(config.WSPort, listenerCh)
		broker.WSTransport = wsTransport
		if err != nil {
			log.Printf("WARN: failed to start WS listener on port %d: %v", config.WSPort, err)
		} else {
			log.Printf("INFO: started WS listener on port %d", config.WSPort)
			hostedServices = append(hostedServices, "ws-listener")
		}
	}
	if config.TLS != nil {
		if config.WSSPort > 0 {
			wssTransport, err := transport.NewWSSTransport(config.WSSPort, config.TLS, listenerCh)
			broker.WSSTransport = wssTransport
			if err != nil {
				log.Printf("WARN: failed to start WSS listener on port %d: %v", config.WSSPort, err)
			} else {
				log.Printf("INFO: started WSS listener on port %d", config.WSSPort)
				hostedServices = append(hostedServices, "wss-listener")
			}
		}
		if config.TLSPort > 0 {
			tlsTransport, err := transport.NewTLSTransport(config.TLSPort, config.TLS, listenerCh)
			broker.TLSTransport = tlsTransport
			if err != nil {
				log.Printf("WARN: failed to start TLS listener on port %d: %v", config.TLSPort, err)
			} else {
				log.Printf("INFO: started TLS listener on port %d", config.TLSPort)
				hostedServices = append(hostedServices, "tls-listener")
			}
		} else {
			log.Printf("WARN: failed to start TLS listener: TLS config not found")
		}
	}
	broker.Listener = l
	if config.NATSURL != "" {
		ch := make(chan STANMessage)
		if err := exportToSTAN(config, ch); err != nil {
			log.Printf("WARN: failed to start STAN message exporter to %s: %v", config.NATSURL, err)
		} else {
			log.Printf("INFO: started NATS message exporter")
			broker.STANOutput = ch
		}
	}
	broker.setupLogs()
	broker.setupSYSTopic()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = os.Getenv("HOSTNAME")
	}
	if hostname == "" {
		hostname = "not_available"
	}
	broker.Peers.Upsert(peers.Peer{
		Metadata: peers.Metadata{
			ID:       broker.ID,
			Hostname: hostname,
			Runtime:  runtime.Version(),
			Services: hostedServices,
			Started:  time.Now().Unix(),
		},
	})
	go broker.oSStatsReporter()
	return broker
}
func (b *Broker) onPeerDown(name string) {
	peer, err := b.Peers.ByID(name)
	if err != nil {
		log.Printf("WARN: received lost event from an unknown peer %s", name)
		return
	}
	log.Printf("WARN: peer %s lost", name)
	b.Peers.Delete(peer.ID)
	set, err := b.Subscriptions.ByPeer(name)
	if err != nil {
		log.Printf("ERR: failed to remove subscriptions from peer %s: %v", name, err)
		return
	}
	set.Apply(func(sub subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})

	sessionSet, err := b.Sessions.ByPeer(name)
	if err != nil {
		log.Printf("ERR: failed to fetch sessions from peer %s: %v", name, err)
		return
	}
	sessionSet.Apply(func(s sessions.Session) {
		if s.WillRetain {
			retainedMessage := topics.RetainedMessage{
				Metadata: topics.Metadata{
					Payload: s.WillPayload,
					Qos:     s.WillQoS,
					Tenant:  s.Tenant,
					Topic:   s.WillTopic,
				},
			}
			b.Topics.Create(retainedMessage)
		}
		recipients, err := b.Subscriptions.ByTopic(s.Tenant, s.WillTopic)
		if err != nil {
			return
		}

		lwt := packet.Publish{
			Payload: s.WillPayload,
			Topic:   s.WillTopic,
			Header: &packet.Header{
				Qos: s.WillQoS,
			},
		}
		recipients.Apply(func(sub subscriptions.Subscription) {
			sub.Sender(lwt)
		})
		b.Sessions.Delete(s.ID, "session_lost")
	})
}

func (b *Broker) Join(hosts []string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		log.Printf("INFO: joining hosts %v", hosts)
		err := b.mesh.Join(hosts)
		if err == nil {
			return
		}
		log.Printf("WARN: failed to join the provided node list: %v", err)
		<-ticker.C
	}
}

func (b *Broker) dispatch(message *rpc.MessagePublished) error {
	session, err := b.Sessions.ByID(message.Recipient)
	if err != nil {
		return err
	}
	packet := packet.Publish{
		Header: &packet.Header{
			Dup:    message.Dup,
			Qos:    message.Qos,
			Retain: message.Retained,
		},
		Payload:   message.Payload,
		Topic:     message.Topic,
		MessageId: 1,
	}
	if session.Transport != nil {
		return session.Transport.Publish(&packet)
	}
	return errors.New("sessio not managed by this node")
}

func (b *Broker) OnBrokerStopped(f func()) func() {
	return b.events.Subscribe("broker_stopped", func(_ events.Event) {
		f()
	})
}
func memUsage() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}
func (b *Broker) oSStatsReporter() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		m := memUsage()
		nbRoutines := runtime.NumGoroutine()
		nbCores := runtime.NumCPU()
		self, err := b.Peers.ByID(b.ID)
		if err != nil {
			return
		}
		self.ComputeUsage = &peers.ComputeUsage{
			Cores:      int64(nbCores),
			Goroutines: int64(nbRoutines),
		}
		self.MemoryUsage = &peers.MemoryUsage{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			NumGC:      m.NumGC,
			Sys:        m.Sys,
		}
		b.Peers.Upsert(self)
		<-ticker.C
	}
}

func (b *Broker) Stop() {
	log.Printf("INFO: stopping Listener aggregator")
	b.Listener.Close()
	log.Printf("INFO: stopping Listener aggregator stopped")
	if b.TCPTransport != nil {
		log.Printf("INFO: stopping TCP listener")
		b.TCPTransport.Close()
		log.Printf("INFO: TCP listener stopped")
	}
	if b.TLSTransport != nil {
		log.Printf("INFO: stopping TLS listener")
		b.TLSTransport.Close()
		log.Printf("INFO: TLS listener stopped")
	}
	if b.WSSTransport != nil {
		log.Printf("INFO: stopping WSS listener")
		b.WSSTransport.Close()
		log.Printf("INFO: WSS listener stopped")
	}
	if b.RPC != nil {
		log.Printf("INFO: stopping RPC listener")
		b.RPC.Close()
		log.Printf("INFO: RPC listener stopped")
	}
}
