package broker

import (
	"log"

	"github.com/vx-labs/mqtt-broker/transport"
)

type SessionConfig struct {
	MaxInflightSize int
}
type Config struct {
	NATSURL    string
	AuthHelper func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error)
	Session    SessionConfig
}

func DefaultConfig() Config {
	log.Print("WARN: Default authentication mecanism is used, therefore all access will be granted")
	return Config{
		Session: SessionConfig{
			MaxInflightSize: 500,
		},
		AuthHelper: func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
			return "_default", nil
		},
	}
}
