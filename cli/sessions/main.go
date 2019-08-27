package main

import (
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/sessions"
	"go.uber.org/zap"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "sessions",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cli.Bootstrap(cmd)
			ctx.AddService(cmd, "sessions", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				return sessions.New(id, logger)
			})
			ctx.Run()
		},
	}
	cli.AddClusterFlags(root)
	cli.AddServiceFlags(root, "sessions")
	root.Execute()
}
