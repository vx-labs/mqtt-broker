package main

import (
	"os"

	"github.com/vx-labs/mqtt-broker/api"
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"go.uber.org/zap"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "api",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cli.Bootstrap(cmd)
			ctx.AddService(cmd, "api", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				tcpPort, _ := cmd.Flags().GetInt("tcp-port")
				tlsPort, _ := cmd.Flags().GetInt("tls-port")

				return api.New(id, logger, mesh, api.Config{
					TcpPort:       tcpPort,
					TlsPort:       tlsPort,
					TlsCommonName: os.Getenv("TLS_CN"),
				})
			})
			ctx.Run()
		},
	}
	cli.AddClusterFlags(root)
	cli.AddServiceFlags(root, "api")
	root.Flags().IntP("tcp-port", "t", 0, "Start TCP listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("tls-port", "s", 0, "Start TLS listener on this port. Specify 0 to disable the listener")
	root.Execute()
}
