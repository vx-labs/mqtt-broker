package identity

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func nomadPort(envkey string) (int, error) {
	port, err := strconv.ParseInt(os.Getenv(envkey), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(port), nil
}

func nomadPublicPort(service string) (int, error) {
	envkey := fmt.Sprintf("NOMAD_HOST_PORT_%s", service)
	return nomadPort(envkey)
}
func nomadPrivatePort(service string) (int, error) {
	envkey := fmt.Sprintf("NOMAD_PORT_%s", service)
	return nomadPort(envkey)
}

func nomadPrivateHost(service string) string {
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
					ip := strings.Split(addresses[0].String(), "/")[0]
					return ip
				}
			}
		}
	}
	panic("could not find a valid network interface")
}
func nomadPublicHost(service string) string {
	envkey := fmt.Sprintf("NOMAD_IP_%s", service)
	return os.Getenv(envkey)
}

// NomadService returns the identity of a service running on Nomad, based on environment variables populated by Nomad
func NomadService(name string) (Identity, error) {
	publicPort, err := nomadPublicPort(name)
	if err != nil {
		return nil, err
	}
	privatePort, err := nomadPrivatePort(name)
	if err != nil {
		return nil, err
	}

	return &identity{
		private: &address{
			host: nomadPrivateHost(name),
			port: privatePort,
		},
		public: &address{
			host: nomadPublicHost(name),
			port: publicPort,
		},
	}, nil
}
