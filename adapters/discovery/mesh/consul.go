package mesh

import (
	fmt "fmt"
	"time"

	consul "github.com/hashicorp/consul/api"

	"go.uber.org/zap"
)

func JoinConsulPeers(api *consul.Client, service string, selfAddress string, selfPort int, mesh *discoveryLayer, logger *zap.Logger) error {
	var index uint64
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		services, meta, err := api.Health().Service(
			service,
			"",
			false,
			&consul.QueryOptions{
				WaitIndex: index,
				WaitTime:  15 * time.Second,
			},
		)
		if err != nil {
			<-ticker.C
			continue
		}
		index = meta.LastIndex
		peers := []string{}
		for _, service := range services {
			if service.Checks.AggregatedStatus() == consul.HealthCritical {
				continue
			}
			if service.Service.Address == selfAddress &&
				service.Service.Port == selfPort {
				continue
			}
			peer := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)
			peers = append(peers, peer)
		}
		if len(peers) > 0 {
			mesh.Join(peers)
		}
	}
}
