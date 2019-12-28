package listener

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"go.uber.org/zap"
)

type Service struct{}

func (s *Service) Name() string {
	return "listener"
}
func (s *Service) Register(cmd *cobra.Command, config *viper.Viper) error {
	cmd.Flags().IntP("listener-tcp-port", "t", 0, "Start TCP listener on this port. Specify 0 to disable the listener")
	config.BindPFlag("listener-tcp-port", cmd.Flags().Lookup("listener-tcp-port"))

	cmd.Flags().IntP("listener-tls-port", "s", 0, "Start TLS listener on this port. Specify 0 to disable the listener")
	config.BindPFlag("listener-tls-port", cmd.Flags().Lookup("listener-tls-port"))

	cmd.Flags().IntP("listener-wss-port", "w", 0, "Start Secure WS listener on this port. Specify 0 to disable the listener")
	config.BindPFlag("listener-wss-port", cmd.Flags().Lookup("listener-wss-port"))

	cmd.Flags().IntP("listener-ws-port", "", 0, "Start WS listener on this port. Specify 0 to disable the listener")
	config.BindPFlag("listener-ws-port", cmd.Flags().Lookup("listener-ws-port"))

	cmd.Flags().StringP("listener-tls-cn", "", "localhost", "Get ACME certificat for this CN")
	config.BindPFlag("listener-tls-cn", cmd.Flags().Lookup("listener-tls-cn"))

	return nil
}

func (s *Service) Run(id string, config *viper.Viper, logger *zap.Logger, mesh cluster.DiscoveryAdapter) cli.Service {
	tcpPort := config.GetInt("listener-tcp-port")
	tlsPort := config.GetInt("listener-tls-port")
	wssPort := config.GetInt("listener-wss-port")
	wsPort := config.GetInt("listener-ws-port")
	cn := config.GetString("listener-tls-cn")
	if cn == "localhost" && os.Getenv("TLS_CN") != "" {
		cn = os.Getenv("TLS_CN")
	}
	return New(id, logger, mesh, Config{
		TCPPort:       tcpPort,
		TLSPort:       tlsPort,
		WSPort:        wsPort,
		WSSPort:       wssPort,
		TLSCommonName: cn,
	})
}
