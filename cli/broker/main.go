package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/transport"
	"go.uber.org/zap"

	"github.com/vx-labs/mqtt-broker/broker"

	"github.com/vx-labs/mqtt-broker/cluster"

	_ "net/http/pprof"

	auth "github.com/vx-labs/iot-mqtt-auth/api"

	"github.com/spf13/cobra"
)

func authHelper(ctx context.Context) func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
	if os.Getenv("BYPASS_AUTH") == "true" {
		return func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
			return "_default", nil
		}
	}
	api, err := auth.New(os.Getenv("AUTH_HOST"))
	if err != nil {
		panic(err)
	}
	return func(transport transport.Metadata, sessionID []byte, username string, password string) (tenant string, err error) {
		success, tenant, err := api.Authenticate(
			ctx,
			auth.WithProtocolContext(
				username,
				password,
			),
			auth.WithTransportContext(transport.Encrypted, transport.RemoteAddress, nil),
		)
		if err != nil {
			log.Printf("ERROR: auth failed: %v", err)
			return "", fmt.Errorf("bad_username_or_password")
		}
		if success {
			return tenant, nil
		}
		return "", fmt.Errorf("bad_username_or_password")
	}
}

func main() {
	root := &cobra.Command{
		Use: "broker",
		Run: func(cmd *cobra.Command, args []string) {

			cli.Run(cmd, "broker", func(id string, logger *zap.Logger, mesh cluster.Mesh) cli.Service {
				config := broker.DefaultConfig()
				if os.Getenv("NOMAD_ALLOC_ID") != "" {
					config.AuthHelper = authHelper(context.Background())
				}
				return broker.New(id, logger, mesh, config)
			})
		},
	}
	cli.AddClusterFlags(root)
	root.Execute()
}
