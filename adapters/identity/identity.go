package identity

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/network"
	"os"
	"strconv"
)

type Identity interface {
	ID() string
	Name() string
	AdvertisedAddress() string
	AdvertisedPort() int
	BindAddress() string
	BindPort() int
}

type Catalog interface {
	Get(name string) Identity
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

func (c *local) Get(name string) Identity {
	config := network.ConfigurationFromFlags(c.v, name)
	return &config
}

type nomad struct {
}

func NewNomadCatalog() Catalog {
	return &nomad{}
}

func (c *nomad) Get(name string) Identity {
	hostPort, err := strconv.ParseInt(os.Getenv(fmt.Sprintf("NOMAD_HOST_PORT_%s", name)), 10, 64)
	if err != nil {
		panic(err)
	}
	bindPort, err := strconv.ParseInt(os.Getenv(fmt.Sprintf("NOMAD_PORT_%s", name)), 10, 64)
	if err != nil {
		panic(err)
	}
	return &nomadIdentity{
		id:                fmt.Sprintf("_nomad-task-%s-%s-%s-%s", os.Getenv("NOMAD_ALLOC_ID"), os.Getenv("NOMAD_TASK_NAME"), name, name),
		name:              name,
		advertisedAddress: os.Getenv(fmt.Sprintf("NOMAD_IP_%s", name)),
		advertisedPort:    int(hostPort),
		bindAddress:       "0.0.0.0",
		bindPort:          int(bindPort),
	}
}
