package providers

import (
	"errors"
	"os"

	"github.com/vx-labs/mqtt-broker/services/auth/pb"
)

var ErrAuthenticationFailed = errors.New("authentication failed")

type Provider interface {
	Authenticate(tenant, id string, p *pb.ProtocolContext, t *pb.TransportContext) bool
}

type denyAllProvider struct{}

func (_ *denyAllProvider) Authenticate(tenant, id string, p *pb.ProtocolContext, t *pb.TransportContext) bool {
	return false
}

var providers map[string]Provider = map[string]Provider{
	"deny": &denyAllProvider{},
	"static-token": &static{
		keys: map[string]string{
			"demo":   "demo",
			"vx:psk": os.Getenv("PSK_PASSWORD"),
		},
	},
}

func Get(id string) Provider {
	v, ok := providers[id]
	if !ok {
		return providers["deny"]
	}
	return v
}
