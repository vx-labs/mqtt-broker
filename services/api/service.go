package api

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"go.uber.org/zap"
)

type Service struct{}

func (s *Service) Name() string {
	return "api"
}
func (s *Service) Register(cmd *cobra.Command, config *viper.Viper) error {
	cmd.Flags().IntP("api-tcp-port", "", 0, "Start API TCP listener on this port. Specify 0 to disable the listener")
	config.BindPFlag("api-tcp-port", cmd.Flags().Lookup("api-tcp-port"))

	cmd.Flags().IntP("api-tls-port", "", 0, "Start API TLS listener on this port. Specify 0 to disable the listener")
	config.BindPFlag("api-tls-port", cmd.Flags().Lookup("api-tls-port"))

	cmd.Flags().StringP("api-tls-cn", "", "localhost", "Get ACME certificat for this CN")
	config.BindPFlag("api-tls-cn", cmd.Flags().Lookup("api-tls-cn"))

	return nil
}
func (s *Service) Run(id string, config *viper.Viper, logger *zap.Logger, mesh cluster.DiscoveryAdapter) cli.Service {
	return New(id, logger, mesh, Config{
		TcpPort:       config.GetInt("api-tcp-port"),
		TlsCommonName: config.GetString("api-tls-cn"),
		TlsPort:       config.GetInt("api-tls-port"),
		logger:        logger,
	})
}
