package nomad

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/hashicorp/consul/api"
	consul "github.com/hashicorp/consul/api"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/listeners"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

type NomadDiscoveryAdapter struct {
	id     string
	api    *consul.Client
	nodeId string
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

func NewNomadDiscoveryAdapter(id string, logger *zap.Logger) *NomadDiscoveryAdapter {
	consulConfig := consul.DefaultConfig()
	consulConfig.HttpClient = http.DefaultClient
	consulAPI, err := consul.NewClient(consulConfig)
	if err != nil {
		logger.Error("failed to connect to consul", zap.Error(err))
		return nil
	}
	resolver.Register(newNomadResolver(consulAPI, logger))
	info, err := consulAPI.Agent().Self()
	if err != nil {
		panic(err)
	}
	return &NomadDiscoveryAdapter{
		id:     id,
		api:    consulAPI,
		nodeId: info["Config"]["NodeID"].(string),
	}
}

func (c *NomadDiscoveryAdapter) EndpointsByService(name string) ([]*pb.NodeService, error) {
	services, _, err := c.api.Health().Service(name, "", false, &api.QueryOptions{AllowStale: false})
	if err != nil {
		return nil, err
	}
	out := make([]*pb.NodeService, len(services))
	for idx, service := range services {
		nomadAllocID := strings.TrimPrefix(service.Service.ID, "_nomad-task-")[0:36]
		out[idx] = &pb.NodeService{
			ID:             service.Service.ID,
			Name:           name,
			Health:         service.Checks.AggregatedStatus(),
			Peer:           nomadAllocID,
			NetworkAddress: fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port),
			Tags:           parseTags(service.Service.Tags),
		}
	}
	return out, nil
}
func (d *NomadDiscoveryAdapter) ListenTCP(id, name string, port int, advertizedAddress string) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	return &listeners.ServiceTCPListener{
		Listener: listener,
	}, nil
}
func (d *NomadDiscoveryAdapter) ListenUDP(id, name string, port int, advertizedAddress string) (net.PacketConn, error) {
	listener, err := net.ListenPacket("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	return &listeners.ServiceUDPListener{
		Listener: listener,
	}, nil
}

func (c *NomadDiscoveryAdapter) AddServiceTag(id, key, value string) error {
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
	_, err = c.api.Catalog().Register(&consul.CatalogRegistration{
		ID:             id,
		Node:           c.nodeId,
		SkipNodeUpdate: true,
		Service: &consul.AgentService{
			Tags: service.Tags,
		},
	}, nil)
	return err
}
func (c *NomadDiscoveryAdapter) RemoveServiceTag(id, key string) error {
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
	_, err = c.api.Catalog().Register(&consul.CatalogRegistration{
		ID:             id,
		SkipNodeUpdate: true,
		Node:           c.nodeId,
		Service: &consul.AgentService{
			Tags: service.Tags,
		},
	}, nil)
	return err
}
func (c *NomadDiscoveryAdapter) DialService(name string, tags ...string) (*grpc.ClientConn, error) {
	key := fmt.Sprintf("nomad:///%s", name)
	if len(tags) > 0 {
		key = fmt.Sprintf("%s?%s", key, strings.Join(tags, "&"))
	}
	return grpc.Dial(key,
		network.GRPCClientOptions()...,
	)
}
func (c *NomadDiscoveryAdapter) Shutdown() error {
	return nil
}
