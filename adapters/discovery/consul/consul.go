package consul

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	consul "github.com/hashicorp/consul/api"
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

func parseTags(tags []string) []*pb.ServiceTag {
	out := make([]*pb.ServiceTag, len(tags))
	for idx, tag := range tags {
		tokens := strings.Split(tag, "=")
		if len(tokens) == 1 {
			out[idx] = &pb.ServiceTag{
				Key: tokens[0],
			}
		} else {
			out[idx] = &pb.ServiceTag{
				Key:   tokens[0],
				Value: tokens[1],
			}
		}
	}
	return out
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

func (c *ConsulDiscoveryAdapter) EndpointsByService(name string) ([]*pb.NodeService, error) {
	services, _, err := c.api.Health().Service(name, "", false, &api.QueryOptions{AllowStale: false})
	if err != nil {
		return nil, err
	}
	out := make([]*pb.NodeService, len(services))
	for idx, service := range services {
		out[idx] = &pb.NodeService{
			ID:             service.Service.ID,
			Name:           name,
			Peer:           service.Service.Meta["node_id"],
			NetworkAddress: fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port),
			Tags:           parseTags(service.Service.Tags),
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
func (c *ConsulDiscoveryAdapter) AddServiceTag(id, key, value string) error {
	service, _, err := c.api.Agent().Service(id, &api.QueryOptions{AllowStale: false})
	if err != nil {
		return err
	}
	updated := false
	for idx, tag := range parseTags(service.Tags) {
		if tag.Key == key {
			if tag.Value == value {
				return nil
			}
			service.Tags[idx] = fmt.Sprintf("%s=%s", tag.Key, value)
			updated = true
		}
	}
	if !updated {
		service.Tags = append(service.Tags, fmt.Sprintf("%s=%s", key, value))
	}
	return c.api.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:                id,
		Name:              service.Service,
		Address:           service.Address,
		Port:              service.Port,
		EnableTagOverride: true,
		Tags:              service.Tags,
		Meta:              service.Meta,
	})
}
func (c *ConsulDiscoveryAdapter) RemoveServiceTag(id, key string) error {
	service, _, err := c.api.Agent().Service(id, &api.QueryOptions{AllowStale: false})
	if err != nil {
		return err
	}
	updated := false
	for idx, tag := range parseTags(service.Tags) {
		if tag.Key == key {
			service.Tags[idx] = service.Tags[len(service.Tags)-1]
			service.Tags = service.Tags[:len(service.Tags)-1]
			updated = true
		}
	}
	if !updated {
		return nil
	}
	return c.api.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:                id,
		Name:              service.Service,
		Address:           service.Address,
		Port:              service.Port,
		EnableTagOverride: true,
		Tags:              service.Tags,
		Meta:              service.Meta,
	})
}
func (c *ConsulDiscoveryAdapter) DialService(name string, tags ...string) (*grpc.ClientConn, error) {
	key := fmt.Sprintf("consul:///%s", name)
	if len(tags) > 0 {
		key = fmt.Sprintf("%s?%s", key, strings.Join(tags, "&"))
	}
	return grpc.Dial(key,
		network.GRPCClientOptions()...,
	)
}
func (c *ConsulDiscoveryAdapter) Shutdown() error {
	return nil
}
func (c *ConsulDiscoveryAdapter) Members() ([]*pb.Peer, error) {
	return nil, errors.New("Unsupported")
}
