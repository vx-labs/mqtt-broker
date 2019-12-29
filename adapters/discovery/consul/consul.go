package consul

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	consul "github.com/hashicorp/consul/api"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	return &ConsulDiscoveryAdapter{
		id:  id,
		api: consulAPI,
	}
}

func (c *ConsulDiscoveryAdapter) EndpointsByService(name string) ([]*pb.NodeService, error) {
	services, _, err := c.api.Health().Service(name, "", false, nil)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.NodeService, len(services))
	for idx, service := range services {
		out[idx] = &pb.NodeService{
			ID:             name,
			Peer:           service.Node.ID,
			NetworkAddress: fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port),
			Tags:           parseTags(service.Service.Tags),
		}
	}
	return out, nil
}
func (c *ConsulDiscoveryAdapter) RegisterService(name, address string) error {
	return errors.New("Unsupported yet")
}
func (c *ConsulDiscoveryAdapter) UnregisterService(name string) error {
	return errors.New("Unsupported yet")
}
func (c *ConsulDiscoveryAdapter) AddServiceTag(service, key, value string) error {
	return errors.New("Unsupported yet")
	/*
		tag := fmt.Sprintf("%s=%s", key, value)
		results, _, err := c.api.Catalog().Service(service, "", nil)
		if err != nil {
			return nil
		}
		for _, result := range results {
			if result.Node == c.id {
				result.ServiceTags = append(result.ServiceTags, tag)
				_, err := c.api.Catalog().Register(&api.CatalogRegistration{
					ID:         result.ID,
					Node:       result.Node,
					NodeMeta:   result.NodeMeta,
					Address:    result.Address,
					Checks:     result.Checks,
					Datacenter: result.Datacenter,
					Service: &api.AgentService{
						ID:      result.ServiceID,
						Address: result.ServiceAddress,
						Tags:    result.ServiceTags,
						Port:    result.ServicePort,
					},
				}, nil)
				return err
			}
		}
		return errors.New("service not found")
	*/
}
func (c *ConsulDiscoveryAdapter) RemoveServiceTag(service, key string) error {
	return errors.New("Unsupported yet")
}
func (c *ConsulDiscoveryAdapter) DialService(name string, tags ...string) (*grpc.ClientConn, error) {
	return nil, errors.New("Unsupported yet")
}
func (c *ConsulDiscoveryAdapter) Shutdown() error {
	return nil
}
func (c *ConsulDiscoveryAdapter) Members() ([]*pb.Peer, error) {
	return nil, errors.New("Unsupported yet")
}
