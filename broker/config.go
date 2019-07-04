package broker

import (
	"crypto/tls"
	"log"

	"github.com/vx-labs/mqtt-broker/broker/transport"
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
	AuthHelper  func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error)
	Session     SessionConfig
}

func DefaultConfig() Config {
	log.Print("WARN: Default authentication mecanism is used, therefore all access will be granted")
	return Config{
		Session: SessionConfig{
			MaxInflightSize: 500,
		},
		RPCPort: 3000,
		AuthHelper: func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
			return "_default", nil
		},
	}
}
