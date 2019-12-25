package broker

import (
	"github.com/vx-labs/mqtt-broker/transport"
)

type SessionConfig struct {
	MaxInflightSize int
}
type Config struct {
	AuthHelper func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error)
	Session    SessionConfig
}

func DefaultConfig() Config {
	return Config{
		Session: SessionConfig{
			MaxInflightSize: 500,
		},
		AuthHelper: func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
			return "_default", nil
		},
	}
}
