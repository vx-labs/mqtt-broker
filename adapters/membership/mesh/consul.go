package mesh

import (
	"io"
	"time"

	"go.uber.org/zap"
)

type joiner interface {
	Join([]string) error
}

func startJoiner(api DiscoveryService, mesh joiner, logger *zap.Logger) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	retries := 10
	for retries > 0 {
		retries--
		services, err := api.DiscoverEndpoints()
		if err == io.EOF {
			logger.Info("not joining membership cluster as no node were provided by discovery", zap.Error(err))
			return nil
		}
		if err != nil {
			<-ticker.C
			continue
		}
		peers := []string{}
		for _, service := range services {
			peers = append(peers, service.NetworkAddress)
		}
		if len(peers) > 0 {
			err := mesh.Join(peers)
			if err == nil {
				logger.Info("membership cluster joined", zap.Error(err))
				return nil
			}
		}
		<-ticker.C
	}
	return nil
}
