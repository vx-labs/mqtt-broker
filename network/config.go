package network

import (
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Configuration struct {
	id                string
	name              string
	tag               string
	advertisedAddress string
	advertisedPort    int
	bindAddress       string
	bindPort          int
}

func (c *Configuration) Name() string {
	return c.name
}
func (c *Configuration) ID() string {
	return c.id
}
func (c *Configuration) Tag() string {
	return c.tag
}
func (c *Configuration) AdvertisedAddress() string {
	return c.advertisedAddress
}
func (c *Configuration) AdvertisedPort() int {
	if c.bindPort == 0 {
		panic("invalid advertised port: 0")
	}
	return c.advertisedPort
}
func (c *Configuration) BindPort() int {
	if c.bindPort == 0 {
		panic("invalid bind port: 0")
	}
	return c.bindPort
}
func (c *Configuration) BindAddress() string {
	return c.bindAddress
}

func randomFreePort(host string) (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", host))
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil

}

func localPrivateHost() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, v := range ifaces {
		if v.Flags&net.FlagLoopback != net.FlagLoopback && v.Flags&net.FlagUp == net.FlagUp {
			h := v.HardwareAddr.String()
			if len(h) == 0 {
				continue
			} else {
				addresses, _ := v.Addrs()
				if len(addresses) > 0 {
					ip := addresses[0]
					if ipnet, ok := ip.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
						if ipnet.IP.To4() != nil {
							return ipnet.IP.String()
						}
					}
				}
			}
		}
	}
	panic("could not find a valid network interface")
}

func advertisedAddressFlagName(name string) string {
	return fmt.Sprintf("%s-advertised-address", name)
}
func advertisedPortFlagName(name string) string {
	return fmt.Sprintf("%s-advertised-port", name)
}
func bindAddressFlagName(name string) string {
	return fmt.Sprintf("%s-bind-address", name)
}
func bindPortFlagName(name string) string {
	return fmt.Sprintf("%s-bind-port", name)
}

func (c Configuration) Describe(name string) string {
	return fmt.Sprintf("INFO: service %s is running on %s:%d and exposed on %s:%d",
		name,
		c.bindAddress, c.bindPort,
		c.advertisedAddress, c.advertisedPort,
	)
}

func ConfigurationFromFlags(v *viper.Viper, name, tag string) Configuration {
	serviceID := fmt.Sprintf("%s-service-id", name)

	config := Configuration{
		id:                v.GetString(serviceID),
		name:              name,
		tag:               tag,
		advertisedAddress: v.GetString(advertisedAddressFlagName(name)),
		advertisedPort:    v.GetInt(advertisedPortFlagName(name)),
		bindAddress:       v.GetString(bindAddressFlagName(name)),
		bindPort:          v.GetInt(bindPortFlagName(name)),
	}

	if len(config.id) == 0 {
		log.Fatalf("empty service id")
	}

	if len(config.advertisedAddress) == 0 {
		config.advertisedAddress = config.bindAddress
	}
	if config.bindPort == 0 {
		randomPort, err := randomFreePort(config.bindAddress)
		if err != nil {
			panic(err)
		}
		config.bindPort = randomPort
	}
	if config.advertisedPort == 0 {
		config.advertisedPort = config.bindPort
	}
	if net.ParseIP(config.bindAddress) == nil {
		log.Fatalf("invalid bind address specified for service %s: %q", name, config.bindAddress)
	}
	if net.ParseIP(config.advertisedAddress) == nil {
		log.Fatalf("invalid advertised address specified for service %s: %q", name, config.advertisedAddress)
	}
	if config.advertisedPort < 1024 || config.advertisedPort > 65535 {
		log.Fatalf("invalid advertised port specified for service %s: %d", name, config.advertisedPort)
	}
	if config.bindPort < 1024 || config.bindPort > 65535 {
		log.Fatalf("invalid bind port specified for service %s: %d", name, config.bindPort)
	}
	return config
}
func RegisterFlagsForService(cmd *cobra.Command, config *viper.Viper, name string, defaultPort int) {
	long := bindPortFlagName(name)
	longAddr := bindAddressFlagName(name)
	advLong := advertisedPortFlagName(name)
	advLongAddr := advertisedAddressFlagName(name)

	defaultAddr := localPrivateHost()
	serviceID := fmt.Sprintf("%s-service-id", name)
	cmd.Flags().StringP(serviceID, "", uuid.New().String(), fmt.Sprintf("%s unique id", name))
	config.BindPFlag(serviceID, cmd.Flags().Lookup(serviceID))

	cmd.Flags().IntP(long, "", defaultPort, fmt.Sprintf("Start %s listener on this port", name))
	config.BindPFlag(long, cmd.Flags().Lookup(long))
	config.BindEnv(long, fmt.Sprintf("NOMAD_PORT_%s", name))

	cmd.Flags().StringP(longAddr, "", defaultAddr, fmt.Sprintf("Start %s listener on this address", name))
	config.BindPFlag(longAddr, cmd.Flags().Lookup(longAddr))

	cmd.Flags().StringP(advLongAddr, "", defaultAddr, fmt.Sprintf("Advertise %s listener on this address", name))
	config.BindPFlag(advLongAddr, cmd.Flags().Lookup(advLongAddr))
	config.BindEnv(advLongAddr, fmt.Sprintf("NOMAD_IP_%s", name))

	cmd.Flags().IntP(advLong, "", 0, fmt.Sprintf("Advertise %s listener on this port", name))
	config.BindPFlag(advLong, cmd.Flags().Lookup(advLong))
	config.BindEnv(advLong, fmt.Sprintf("NOMAD_HOST_PORT_%s", name))

}
