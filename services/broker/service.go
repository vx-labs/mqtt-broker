package broker

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/transport"
	"go.uber.org/zap"
)

type Service struct{}

func (s *Service) Name() string {
	return "broker"
}
func (s *Service) Register(cmd *cobra.Command, config *viper.Viper) error {

	return nil
}
func (s *Service) Run(id string, config *viper.Viper, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
	return New(id, logger, mesh, Config{
		AuthHelper: func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
			if username == "vx:psk" || username == "vx-psk" {
				if password == os.Getenv("PSK_PASSWORD") {
					return "_default", nil
				}
			}
			return "", fmt.Errorf("bad_username_or_password")
		},
		Session: SessionConfig{
			MaxInflightSize: 20, // not used for now
		},
	})
}
