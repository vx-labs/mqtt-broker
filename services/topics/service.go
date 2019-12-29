package topics

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/cli"
	"go.uber.org/zap"
)

type Service struct{}

func (s *Service) Name() string {
	return "topics"
}
func (s *Service) Register(cmd *cobra.Command, config *viper.Viper) error {
	return nil
}
func (s *Service) Run(id string, config *viper.Viper, logger *zap.Logger, mesh discovery.DiscoveryAdapter) cli.Service {
	return New(id, logger)
}
