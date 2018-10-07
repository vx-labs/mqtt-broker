package identity

import (
	"fmt"
	"net"
	"strings"
)

func getRandomPort(host string) (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", host))
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
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
					ip := strings.Split(addresses[0].String(), "/")[0]
					return ip
				}
			}
		}
	}
	panic("could not find a valid network interface")
}

func LocalService() (Identity, error) {
	host := localPrivateHost()
	port, err := getRandomPort(host)
	if err != nil {
		return nil, err
	}
	return &identity{
		private: &address{
			host: host,
			port: port,
		},
		public: &address{
			host: host,
			port: port,
		},
	}, nil
}
