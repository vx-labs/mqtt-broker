package broker

import (
	"crypto/tls"
	"log"

	"github.com/vx-labs/mqtt-broker/broker/listener"
)

type Config struct {
	TCPPort    int
	TLS        *tls.Config
	TLSPort    int
	WSSPort    int
	GossipPort int
	AuthHelper func(transport listener.Transport, sessionID, username string, password string) (tenant string, id string, err error)
}

func DefaultConfig() Config {
	return Config{
		AuthHelper: func(transport listener.Transport, sessionID, username string, password string) (tenant string, id string, err error) {
			log.Print("WARN: Default authentication mecanism is used, therefore all access will be granted")
			return "_default", sessionID, nil
		},
	}
}
