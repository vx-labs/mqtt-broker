package main

import (
	"strconv"
	"strings"

	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/messages"
	"go.uber.org/zap"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "messages",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cli.Bootstrap(cmd)
			ctx.AddService(cmd, "messages", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				flagConfig, _ := cmd.Flags().GetStringSlice("initial-stream-config")
				serverConfig := messages.ServerConfig{}
				for _, element := range flagConfig {
					tokens := strings.Split(element, ":")
					if len(tokens) != 2 {
						panic("invalid stream config provided")
					}
					shardCount, err := strconv.ParseInt(tokens[1], 10, 64)
					if err != nil {
						logger.Fatal("failed to parse stream config", zap.Error(err))
					}
					serverConfig.InitialsStreams = append(serverConfig.InitialsStreams, messages.ServerStreamConfig{
						ID:         tokens[0],
						ShardCount: shardCount,
					})
				}
				return messages.New(id, serverConfig, logger)
			})
			ctx.Run()
		},
	}
	cli.AddClusterFlags(root)
	cli.AddServiceFlags(root, "messages")
	root.Flags().StringSlice("initial-stream-config", nil, "Create this stream at startup if it does not exist. Syntax is stream-id:shard-count. Ex: --initial-stream-config='messages:3'")

	root.Execute()
}
