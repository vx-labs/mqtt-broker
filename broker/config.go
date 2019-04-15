package broker

import (
	"crypto/tls"
	"log"

	"github.com/vx-labs/mqtt-broker/broker/listener"
	"github.com/vx-labs/mqtt-broker/identity"
)

type SessionConfig struct {
	MaxInflightSize int
}
type Config struct {
	TCPPort     int
	TLS         *tls.Config
	TLSPort     int
	WSSPort     int
	WSPort      int
	RPCPort     int
	RPCIdentity identity.Identity
	GossipPort  int
	NATSURL     string
	AuthHelper  func(transport listener.Transport, sessionID []byte, username string, password string) (tenant string, err error)
	Session     SessionConfig
}

func DefaultConfig() Config {
	return Config{
		Session: SessionConfig{
			MaxInflightSize: 500,
		},
		RPCPort: 0,
		AuthHelper: func(transport listener.Transport, sessionID []byte, username string, password string) (tenant string, err error) {
			log.Print("WARN: Default authentication mecanism is used, therefore all access will be granted")
			return "_default", nil
		},
	}
}
