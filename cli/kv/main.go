package main

import (
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/kv"
	"go.uber.org/zap"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "kv",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cli.Bootstrap(cmd)
			ctx.AddService(cmd, "kv", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				return kv.New(id, logger)
			})
			ctx.Run()
		},
	}
	cli.AddClusterFlags(root)
	cli.AddServiceFlags(root, "kv")
	root.Execute()
}
