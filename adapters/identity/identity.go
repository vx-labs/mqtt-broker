package identity

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/network"
)

type Identity interface {
	ID() string
	Name() string
	Tag() string
	AdvertisedAddress() string
	AdvertisedPort() int
	BindAddress() string
	BindPort() int
}

type Catalog interface {
	Get(name, tag string) Identity
}

type local struct {
	v    *viper.Viper
	data map[string]Identity
}

func NewCatalog(v *viper.Viper) Catalog {
	return &local{
		v:    v,
		data: map[string]Identity{},
	}
}

func (c *local) Get(name, tag string) Identity {
	config := network.ConfigurationFromFlags(c.v, name, tag)
	return &config
}

type nomad struct {
}

func NewNomadCatalog() Catalog {
	return &nomad{}
}

func (c *nomad) Get(name, tag string) Identity {
	value := os.Getenv(fmt.Sprintf("NOMAD_HOST_PORT_%s", tag))
	if value == "" {
		return nil
	}
	hostPort, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("service %s: failed to parse NOMAD_HOST_PORT_%s into an integer: %v", name, tag, err))
	}
	bindPort, err := strconv.ParseInt(os.Getenv(fmt.Sprintf("NOMAD_PORT_%s", tag)), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("service %s: failed to parse NOMAD_PORT_%s into an integer: %v", name, tag, err))
	}
	return &nomadIdentity{
		id:                fmt.Sprintf("_nomad-task-%s-%s-%s-%s", os.Getenv("NOMAD_ALLOC_ID"), os.Getenv("NOMAD_TASK_NAME"), name, tag),
		name:              name,
		tag:               tag,
		advertisedAddress: os.Getenv(fmt.Sprintf("NOMAD_IP_%s", name)),
		advertisedPort:    int(hostPort),
		bindAddress:       "0.0.0.0",
		bindPort:          int(bindPort),
	}
}
