package messages

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/cli"
	"go.uber.org/zap"
)

type Service struct{}

func (s *Service) Name() string {
	return "messages"
}
func (s *Service) Register(cmd *cobra.Command, config *viper.Viper) error {
	cmd.Flags().StringSlice("initial-stream-config", nil, "Create this stream at startup if it does not exist. Syntax is stream-id:shard-count. Ex: --initial-stream-config='messages:3'")
	config.BindPFlag("initial-stream-config", cmd.Flags().Lookup("initial-stream-config"))
	return nil
}
func (s *Service) Run(id string, config *viper.Viper, logger *zap.Logger, mesh discovery.DiscoveryAdapter) cli.Service {
	flagConfig := config.GetStringSlice("initial-stream-config")
	serverConfig := ServerConfig{}
	for _, element := range flagConfig {
		tokens := strings.Split(element, ":")
		if len(tokens) != 2 {
			panic("invalid stream config provided")
		}
		shardCount, err := strconv.ParseInt(tokens[1], 10, 64)
		if err != nil {
			logger.Fatal("failed to parse stream config", zap.Error(err))
		}
		serverConfig.InitialsStreams = append(serverConfig.InitialsStreams, ServerStreamConfig{
			ID:         tokens[0],
			ShardCount: shardCount,
		})
	}
	return New(id, serverConfig, logger)
}
