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
	nodeID string
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
		nodeID: info["Config"]["NodeID"].(string),
	}
}

func (c *NomadDiscoveryAdapter) EndpointsByService(name, tag string) ([]*pb.NodeService, error) {
	services, _, err := c.api.Health().Service(name, tag, false, &api.QueryOptions{AllowStale: false})
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
			Tag:            tag,
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

func (c *NomadDiscoveryAdapter) DialService(name string, tag string) (*grpc.ClientConn, error) {
	key := fmt.Sprintf("nomad:///%s", name)
	key = fmt.Sprintf("%s?tag=%s", key, tag)

	return grpc.Dial(key,
		network.GRPCClientOptions()...,
	)
}
func (c *NomadDiscoveryAdapter) Shutdown() error {
	return nil
}
