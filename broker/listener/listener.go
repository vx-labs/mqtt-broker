package listener

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/google/btree"
	"github.com/vx-labs/mqtt-broker/broker/listener/transport"
	"github.com/vx-labs/mqtt-protocol/encoder"
	"github.com/vx-labs/mqtt-protocol/packet"
)

var (
	ErrSessionNotFound = errors.New("session not found on this endpoint")
)

type Broker interface {
	Connect(transport.Metadata, *packet.Connect) (string, *packet.ConnAck, error)
	Disconnect(string, *packet.Disconnect) error
	Publish(string, *packet.Publish) (*packet.PubAck, error)
	Subscribe(string, *packet.Subscribe) (*packet.SubAck, error)
	Unsubscribe(string, *packet.Unsubscribe) (*packet.UnsubAck, error)
	RemoveSession(string) error
}

type Endpoint interface {
	Publish(id string, publish *packet.Publish) error
	CloseSession(string) error
	Close() error
}

type endpoint struct {
	mutex      sync.Mutex
	sessions   *btree.BTree
	transports []net.Listener
	broker     Broker
}

func (local *endpoint) newSession(metadata transport.Metadata) error {
	go local.runLocalSession(metadata)
	return nil
}
func (local *endpoint) Close() error {
	for idx := range local.transports {
		local.transports[idx].Close()
	}
	return nil
}
func (local *endpoint) CloseSession(id string) error {
	local.mutex.Lock()
	defer local.mutex.Unlock()
	session := local.sessions.Delete(&localSession{
		id: id,
	})
	if session != nil {
		return session.(*localSession).transport.Close()
	}
	return ErrSessionNotFound
}
func (local *endpoint) Publish(id string, publish *packet.Publish) error {
	session := local.sessions.Get(&localSession{
		id: id,
	})
	if session == nil {
		return ErrSessionNotFound
	}
	return session.(*localSession).encoder.Publish(publish)
}

type localSession struct {
	id        string
	encoder   *encoder.Encoder
	transport io.Closer
	quit      chan struct{}
}

func (local *localSession) Less(remote btree.Item) bool {
	return strings.Compare(local.id, remote.(*localSession).id) > 0
}

type Config struct {
	TCPPort int
	TLS     *tls.Config
	TLSPort int
	WSSPort int
	WSPort  int
}

func New(broker Broker, config Config) Endpoint {
	local := &endpoint{
		broker:   broker,
		sessions: btree.New(2),
	}
	if config.TCPPort > 0 {
		tcpTransport, err := transport.NewTCPTransport(config.TCPPort, local.newSession)
		if err != nil {
			log.Printf("WARN: failed to start TCP listener on port %d: %v", config.TCPPort, err)
		} else {
			local.transports = append(local.transports, tcpTransport)
		}
	}
	if config.WSPort > 0 {
		wsTransport, err := transport.NewWSTransport(config.WSPort, local.newSession)
		if err != nil {
			log.Printf("WARN: failed to start WS listener on port %d: %v", config.WSPort, err)
		} else {
			log.Printf("INFO: started WS listener on port %d", config.WSPort)
			local.transports = append(local.transports, wsTransport)
		}
	}
	if config.TLS != nil {
		if config.WSSPort > 0 {
			wssTransport, err := transport.NewWSSTransport(config.WSSPort, config.TLS, local.newSession)
			if err != nil {
				log.Printf("WARN: failed to start WSS listener on port %d: %v", config.WSSPort, err)
			} else {
				log.Printf("INFO: started WSS listener on port %d", config.WSSPort)
				local.transports = append(local.transports, wssTransport)
			}
		}
		if config.TLSPort > 0 {
			tlsTransport, err := transport.NewTLSTransport(config.TLSPort, config.TLS, local.newSession)
			if err != nil {
				log.Printf("WARN: failed to start TLS listener on port %d: %v", config.TLSPort, err)
			} else {
				log.Printf("INFO: started TLS listener on port %d", config.TLSPort)
				local.transports = append(local.transports, tlsTransport)
			}
		}
	}
	return local
}
