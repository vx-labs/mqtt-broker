package queues

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"go.uber.org/zap"
)

type Service struct{}

func (s *Service) Name() string {
	return "queues"
}
func (s *Service) Register(cmd *cobra.Command, config *viper.Viper) error {
	return nil
}
func (s *Service) Run(id string, config *viper.Viper, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
	return New(id, logger)
}
