package broker

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"

	"github.com/vx-labs/mqtt-broker/events"
	"github.com/vx-labs/mqtt-broker/peers"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/topics"

	"github.com/vx-labs/mqtt-broker/broker/listener/transport"
	"github.com/vx-labs/mqtt-broker/broker/rpc"

	"github.com/weaveworks/mesh"

	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/vx-labs/mqtt-broker/broker/listener"

	"github.com/golang/protobuf/proto"

	"github.com/vx-labs/mqtt-broker/identity"

	"github.com/vx-labs/mqtt-broker/broker/peer"
	"github.com/vx-labs/mqtt-broker/subscriptions"
)

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/broker/ --go_out=plugins=grpc:. events.proto

type PeerStore interface {
	ByID(id string) (*peers.Peer, error)
	All() (peers.PeerList, error)
	ByMeshID(id uint64) (*peers.Peer, error)
	Upsert(sess *peers.Peer) error
	Delete(id string) error
	On(event string, handler func(*peers.Peer)) func()
}
type SessionStore interface {
	ByID(id string) (*sessions.Session, error)
	ByPeer(peer uint64) (sessions.SessionList, error)
	All() (sessions.SessionList, error)
	Exists(id string) bool
	Upsert(sess *sessions.Session) error
	Delete(id, reason string) error
	On(event string, handler func(*sessions.Session)) func()
}

type TopicStore interface {
	Create(message *topics.RetainedMessage) error
	ByTopicPattern(tenant string, pattern []byte) (topics.RetainedMessageList, error)
	All() (topics.RetainedMessageList, error)
	On(event string, handler func(*topics.RetainedMessage)) func()
}
type SubscriptionStore interface {
	ByTopic(tenant string, pattern []byte) (*subscriptions.SubscriptionList, error)
	ByID(id string) (*subscriptions.Subscription, error)
	All() (subscriptions.SubscriptionList, error)
	ByPeer(peer uint64) (subscriptions.SubscriptionList, error)
	BySession(id string) (subscriptions.SubscriptionList, error)
	Sessions() ([]string, error)
	Create(subscription *subscriptions.Subscription) error
	Delete(id string) error
	On(event string, handler func(*subscriptions.Subscription)) func()
}
type Broker struct {
	ID            string
	authHelper    func(transport listener.Transport, sessionID, username string, password string) (tenant string, id string, err error)
	Peer          *peer.Peer
	Subscriptions SubscriptionStore
	Sessions      SessionStore
	Topics        TopicStore
	Peers         PeerStore
	events        *events.Bus
	Listener      io.Closer
	TCPTransport  io.Closer
	TLSTransport  io.Closer
	WSSTransport  io.Closer
	RPC           io.Closer
}

func New(id identity.Identity, config Config) *Broker {
	broker := &Broker{
		authHelper: config.AuthHelper,
		events:     events.NewEventBus(),
	}
	hostedServices := []string{}
	broker.Peer = peer.NewPeer(id, broker.onPeerDown, broker.onUnicast)
	subscriptionsStore, err := subscriptions.NewMemDBStore(broker.Peer.Router())
	if err != nil {
		log.Fatal(err)
	}
	hostedServices = append(hostedServices, "subscriptions-store")
	peersStore, err := peers.NewPeerStore(broker.Peer.Router())
	if err != nil {
		log.Fatal(err)
	}
	hostedServices = append(hostedServices, "peers-store")
	topicssStore, err := topics.NewMemDBStore(broker.Peer.Router())
	if err != nil {
		log.Fatal(err)
	}
	hostedServices = append(hostedServices, "topics-store")
	sessionsStore, err := sessions.NewSessionStore(broker.Peer.Router())
	if err != nil {
		log.Fatal(err)
	}
	hostedServices = append(hostedServices, "sessions-store")
	broker.Peers = peersStore
	broker.Topics = topicssStore
	broker.Subscriptions = subscriptionsStore
	broker.Sessions = sessionsStore

	l, listenerCh := listener.New(broker, config.Session.MaxInflightSize)
	if config.RPCPort > 0 {
		broker.RPC = rpc.New(config.RPCPort, broker)
		hostedServices = append(hostedServices, "rpc-listener")
	}
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
	if config.TLS != nil {
		if config.WSSPort > 0 {
			wssTransport, err := transport.NewWSSTransport(config.WSSPort, config.TLS, listenerCh)
			broker.WSSTransport = wssTransport
			if err != nil {
				log.Printf("WARN: failed to start WSS listener on port %d: %v", config.TLSPort, err)
			} else {
				log.Printf("INFO: started WSS listener on port %d", config.TLSPort)
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
	broker.setupLogs()
	broker.setupSYSTopic()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = os.Getenv("HOSTNAME")
	}
	if hostname == "" {
		hostname = "not_available"
	}
	broker.ID = uuid.New().String()
	broker.Peers.Upsert(&peers.Peer{
		ID:       broker.ID,
		MeshID:   uint64(broker.Peer.Name()),
		Hostname: hostname,
		Runtime:  runtime.Version(),
		Services: hostedServices,
	})
	go broker.oSStatsReporter()
	broker.Peer.Start()
	return broker
}
func (b *Broker) onUnicast(payload []byte) {
	message := &MessagePublished{}
	err := proto.Unmarshal(payload, message)
	if err != nil {
		return
	}
	b.dispatch(message)
}
func (b *Broker) resolveRecipients(tenant string, topic []byte, defaultQoS int32) ([]string, []int32) {
	recipients, err := b.Subscriptions.ByTopic(tenant, topic)
	if err != nil {
		return []string{}, []int32{}
	}
	set := make([]string, 0, len(recipients.Subscriptions))
	qosSet := make([]int32, 0, len(recipients.Subscriptions))
	recipients.Apply(func(s *subscriptions.Subscription) {
		set = append(set, s.SessionID)
		qos := defaultQoS
		if qos > s.Qos {
			qos = s.Qos
		}
		qosSet = append(qosSet, qos)
	})
	return set, qosSet
}
func (b *Broker) onPeerDown(name mesh.PeerName) {
	peer, err := b.Peers.ByMeshID(uint64(name))
	if err != nil {
		log.Printf("WARN: received lost event from an unknown peer %s", name.String())
		return
	}
	b.Peers.Delete(peer.ID)
	set, err := b.Subscriptions.ByPeer(uint64(name))
	if err != nil {
		log.Printf("ERR: failed to remove subscriptions from peer %d: %v", name, err)
		return
	}
	set.Apply(func(sub *subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
	})

	sessionSet, err := b.Sessions.ByPeer(uint64(name))
	if err != nil {
		log.Printf("ERR: failed to fetch sessions from peer %d: %v", name, err)
		return
	}
	sessionSet.Apply(func(s *sessions.Session) {
		if s.WillRetain {
			retainedMessage := &topics.RetainedMessage{
				Payload: s.WillPayload,
				Qos:     s.WillQoS,
				Tenant:  s.Tenant,
				Topic:   s.WillTopic,
			}
			b.Topics.Create(retainedMessage)
		}
		recipients, err := b.Subscriptions.ByTopic(s.Tenant, s.WillTopic)
		if err != nil {
			return
		}

		message := &MessagePublished{
			Payload:   s.WillPayload,
			Topic:     s.WillTopic,
			Qos:       make([]int32, 0, len(recipients.Subscriptions)),
			Recipient: make([]string, 0, len(recipients.Subscriptions)),
		}
		recipients.Apply(func(sub *subscriptions.Subscription) {
			message.Recipient = append(message.Recipient, sub.SessionID)
			qos := s.WillQoS
			if qos > sub.Qos {
				qos = sub.Qos
			}
			message.Qos = append(message.Qos, qos)
		})
		b.dispatch(message)
		b.Sessions.Delete(s.ID, "peer_lost")
	})
}

func (b *Broker) Join(hosts []string) {
	log.Printf("INFO: joining hosts %v", hosts)
	b.Peer.Join(hosts)
}

func (b *Broker) dispatch(message *MessagePublished) {
	for idx, recipient := range message.Recipient {
		packet := &packet.Publish{
			Header: &packet.Header{
				Dup:    message.Dup,
				Qos:    message.Qos[idx],
				Retain: message.Retained,
			},
			Payload:   message.Payload,
			Topic:     message.Topic,
			MessageId: 1,
		}
		b.events.Emit(events.Event{
			Key:   fmt.Sprintf("message_published/%s", recipient),
			Entry: packet,
		})
	}
}

func (b *Broker) OnMessagePublished(recipient string, f func(p *packet.Publish)) func() {
	return b.events.Subscribe(fmt.Sprintf("message_published/%s", recipient), func(ev events.Event) {
		f(ev.Entry.(*packet.Publish))
	})

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
