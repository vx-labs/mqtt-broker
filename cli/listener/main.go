package main

import (
	"os"

	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/listener"
	"go.uber.org/zap"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "listener",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Run(cmd, "listener", func(id string, logger *zap.Logger, mesh cluster.Mesh) cli.Service {
				tcpPort, _ := cmd.Flags().GetInt("tcp-port")
				tlsPort, _ := cmd.Flags().GetInt("tls-port")
				wssPort, _ := cmd.Flags().GetInt("wss-port")
				wsPort, _ := cmd.Flags().GetInt("ws-port")
				cn, _ := cmd.Flags().GetString("tls-cn")
				if cn == "" {
					cn = os.Getenv("TLS_CN")
				}
				return listener.New(id, logger, mesh, listener.Config{
					TCPPort:       tcpPort,
					TLSPort:       tlsPort,
					WSPort:        wsPort,
					WSSPort:       wssPort,
					TLSCommonName: cn,
				})
			})
		},
	}
	cli.AddClusterFlags(root)
	root.Flags().IntP("tcp-port", "t", 0, "Start TCP listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("tls-port", "s", 0, "Start TLS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("wss-port", "w", 0, "Start Secure WS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("ws-port", "", 0, "Start WS listener on this port. Specify 0 to disable the listener")
	root.Flags().StringP("tls-cn", "", "localhost", "Get ACME certificat for this CN")
	root.Execute()
}
