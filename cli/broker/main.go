package main

import (
	"context"
	"fmt"
	"os"

	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/transport"
	"go.uber.org/zap"

	"github.com/vx-labs/mqtt-broker/broker"

	"github.com/vx-labs/mqtt-broker/cluster"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
)

func authHelper(ctx context.Context) func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
	if os.Getenv("BYPASS_AUTH") == "true" {
		return func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
			return "_default", nil
		}
	}
	return func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
		if username == "vx:psk" || username == "vx-psk" {
			if password == os.Getenv("PSK_PASSWORD") {
				return "_default", nil
			}
		}
		return "", fmt.Errorf("bad_username_or_password")
	}
}

func main() {
	root := &cobra.Command{
		Use: "broker",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cli.Bootstrap(cmd)
			ctx.AddService(cmd, "broker", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				config := broker.DefaultConfig()
				if os.Getenv("NOMAD_ALLOC_ID") != "" {
					config.AuthHelper = authHelper(context.Background())
				}
				return broker.New(id, logger, mesh, config)
			})
			ctx.Run()
		},
	}
	cli.AddClusterFlags(root)
	cli.AddServiceFlags(root, "broker")
	root.Execute()
}
