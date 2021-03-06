package consul

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/hashicorp/consul/api"
	consul "github.com/hashicorp/consul/api"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/listeners"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

type ConsulDiscoveryAdapter struct {
	id  string
	api *consul.Client
}

func NewConsulDiscoveryAdapter(id string, logger *zap.Logger) *ConsulDiscoveryAdapter {
	consulConfig := consul.DefaultConfig()
	consulConfig.HttpClient = http.DefaultClient
	consulAPI, err := consul.NewClient(consulConfig)
	if err != nil {
		logger.Error("failed to connect to consul", zap.Error(err))
		return nil
	}
	resolver.Register(newConsulResolver(consulAPI, logger))
	return &ConsulDiscoveryAdapter{
		id:  id,
		api: consulAPI,
	}
}

func (c *ConsulDiscoveryAdapter) EndpointsByService(name, tag string) ([]*pb.NodeService, error) {
	services, _, err := c.api.Health().Service(name, tag, false, &api.QueryOptions{AllowStale: false})
	if err != nil {
		return nil, err
	}
	out := make([]*pb.NodeService, len(services))
	for idx, service := range services {
		out[idx] = &pb.NodeService{
			ID:             service.Service.ID,
			Name:           name,
			Health:         service.Checks.AggregatedStatus(),
			Peer:           service.Service.Meta["node_id"],
			NetworkAddress: fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port),
			Tag:            tag,
		}
	}
	return out, nil
}
func (c *ConsulDiscoveryAdapter) RegisterTCPService(id, name, address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	intPort, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return err
	}
	return c.api.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: host,
		Port:    int(intPort),
		Meta: map[string]string{
			"node_id": c.id,
		},
		EnableTagOverride: true,
		Check: &api.AgentServiceCheck{
			CheckID:                        fmt.Sprintf("check-tcp-%s-%s", name, id),
			Name:                           fmt.Sprintf("TCP Check on address %s", address),
			DeregisterCriticalServiceAfter: "5m",
			TCP:                            address,
			Interval:                       "10s",
			Timeout:                        "2s",
		},
	})
}
func (d *ConsulDiscoveryAdapter) ListenTCP(id, name string, port int, advertizedAddress string) (net.Listener, error) {
	err := d.RegisterTCPService(id, name, advertizedAddress)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	return &listeners.ServiceTCPListener{
		CloseCallback: func() error {
			return d.UnregisterService(id)
		},
		Listener: listener,
	}, nil
}
func (c *ConsulDiscoveryAdapter) RegisterUDPService(id, name, address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	intPort, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return err
	}
	return c.api.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: host,
		Port:    int(intPort),
		Meta: map[string]string{
			"node_id": c.id,
		},
		EnableTagOverride: true,
	})
}

func (d *ConsulDiscoveryAdapter) ListenUDP(id, name string, port int, advertizedAddress string) (net.PacketConn, error) {
	err := d.RegisterUDPService(id, name, advertizedAddress)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenPacket("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	return &listeners.ServiceUDPListener{
		CloseCallback: func() error {
			return d.UnregisterService(id)
		},
		Listener: listener,
	}, nil
}

func (c *ConsulDiscoveryAdapter) RegisterGRPCService(id, name, address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	intPort, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return err
	}
	return c.api.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: host,
		Port:    int(intPort),
		Meta: map[string]string{
			"node_id": c.id,
		},
		EnableTagOverride: true,
		Check: &api.AgentServiceCheck{
			CheckID:                        fmt.Sprintf("check-grpc-%s-%s", name, id),
			Name:                           fmt.Sprintf("GRPC Check on address %s", address),
			DeregisterCriticalServiceAfter: "5m",
			GRPC:                           address,
			GRPCUseTLS:                     true,
			Interval:                       "10s",
			Timeout:                        "2s",
		},
	})
}
func (c *ConsulDiscoveryAdapter) UnregisterService(id string) error {
	return c.api.Agent().ServiceDeregister(id)
}

func (c *ConsulDiscoveryAdapter) DialService(name string, tag string) (*grpc.ClientConn, error) {
	key := fmt.Sprintf("consul:///%s", name)
	key = fmt.Sprintf("%s?tag=%s", key, tag)

	return grpc.Dial(key,
		network.GRPCClientOptions()...,
	)
}
func (c *ConsulDiscoveryAdapter) Shutdown() error {
	return nil
}
