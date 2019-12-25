package network

import (
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Configuration struct {
	AdvertisedAddress string
	AdvertisedPort    int
	BindAddress       string
	BindPort          int
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
		c.BindAddress, c.BindPort,
		c.AdvertisedAddress, c.AdvertisedPort,
	)
}

func ConfigurationFromFlags(cmd *cobra.Command, v *viper.Viper, name string) Configuration {
	config := Configuration{
		AdvertisedAddress: v.GetString(advertisedAddressFlagName(name)),
		AdvertisedPort:    v.GetInt(advertisedPortFlagName(name)),
		BindAddress:       v.GetString(bindAddressFlagName(name)),
		BindPort:          v.GetInt(bindPortFlagName(name)),
	}

	if len(config.AdvertisedAddress) == 0 {
		config.AdvertisedAddress = config.BindAddress
	}
	if config.BindPort == 0 {
		randomPort, err := randomFreePort(config.BindAddress)
		if err != nil {
			panic(err)
		}
		config.BindPort = randomPort
	}
	if config.AdvertisedPort == 0 {
		config.AdvertisedPort = config.BindPort
	}
	if net.ParseIP(config.BindAddress) == nil {
		log.Fatalf("invalid bind address specified for service %s: %q", name, config.BindAddress)
	}
	if net.ParseIP(config.AdvertisedAddress) == nil {
		log.Fatalf("invalid advertised address specified for service %s: %q", name, config.AdvertisedAddress)
	}
	if config.AdvertisedPort < 1024 || config.AdvertisedPort > 65535 {
		log.Fatalf("invalid advertised port specified for service %s: %d", name, config.AdvertisedPort)
	}
	if config.BindPort < 1024 || config.BindPort > 65535 {
		log.Fatalf("invalid bind port specified for service %s: %d", name, config.BindPort)
	}
	return config
}
func RegisterFlagsForService(cmd *cobra.Command, config *viper.Viper, name string, defaultPort int) {
	long := bindPortFlagName(name)
	longAddr := bindAddressFlagName(name)
	advLong := advertisedPortFlagName(name)
	advLongAddr := advertisedAddressFlagName(name)

	defaultAddr := localPrivateHost()

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
