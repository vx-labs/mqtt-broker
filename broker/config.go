package broker

import (
	"crypto/tls"
	"log"

	"github.com/vx-labs/mqtt-broker/broker/listener"
)

type SessionConfig struct {
	MaxInflightSize int
}
type Config struct {
	TCPPort    int
	TLS        *tls.Config
	TLSPort    int
	WSSPort    int
	WSPort     int
	RPCPort    int
	GossipPort int
	AuthHelper func(transport listener.Transport, sessionID, username string, password string) (tenant string, id string, err error)
	Session    SessionConfig
}

func DefaultConfig() Config {
	return Config{
		Session: SessionConfig{
			MaxInflightSize: 100,
		},
		RPCPort: 9090,
		AuthHelper: func(transport listener.Transport, sessionID, username string, password string) (tenant string, id string, err error) {
			log.Print("WARN: Default authentication mecanism is used, therefore all access will be granted")
			return "_default", sessionID, nil
		},
	}
}
