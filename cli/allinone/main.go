package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vx-labs/mqtt-broker/api"
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/kv"
	"github.com/vx-labs/mqtt-broker/listener"
	"github.com/vx-labs/mqtt-broker/messages"
	"github.com/vx-labs/mqtt-broker/queues"
	"github.com/vx-labs/mqtt-broker/router"
	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
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
			ctx := cli.Bootstrap(cmd)
			ctx.AddService(cmd, "broker", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				config := broker.DefaultConfig()
				if os.Getenv("NOMAD_ALLOC_ID") != "" {
					config.AuthHelper = authHelper(context.Background())
				}
				return broker.New(id, logger, mesh, config)
			})
			ctx.AddService(cmd, "api", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				tcpPort, _ := cmd.Flags().GetInt("api-tcp-port")
				tlsPort, _ := cmd.Flags().GetInt("api-tls-port")

				return api.New(id, logger, mesh, api.Config{
					TcpPort:       tcpPort,
					TlsPort:       tlsPort,
					TlsCommonName: os.Getenv("TLS_CN"),
				})
			})
			ctx.AddService(cmd, "listener", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				tcpPort, _ := cmd.Flags().GetInt("tcp-port")
				tlsPort, _ := cmd.Flags().GetInt("tls-port")
				wssPort, _ := cmd.Flags().GetInt("wss-port")
				wsPort, _ := cmd.Flags().GetInt("ws-port")
				cn, _ := cmd.Flags().GetString("tls-cn")
				if cn == "localhost" && os.Getenv("TLS_CN") != "" {
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
			ctx.AddService(cmd, "sessions", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				return sessions.New(id, logger)
			})
			ctx.AddService(cmd, "queues", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				return queues.New(id, logger)
			})
			ctx.AddService(cmd, "messages", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				return messages.New(id, logger)
			})
			ctx.AddService(cmd, "kv", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				return kv.New(id, logger)
			})
			ctx.AddService(cmd, "subscriptions", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
				return subscriptions.New(id, logger)
			})
			if withRouter, _ := cmd.Flags().GetBool("with-router-cn"); withRouter {
				ctx.AddService(cmd, "router", func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) cli.Service {
					return router.New(id, logger, mesh)
				})
			}

			ctx.Run()
		},
	}
	cli.AddClusterFlags(root)
	cli.AddServiceFlags(root, "broker")
	cli.AddServiceFlags(root, "api")
	cli.AddServiceFlags(root, "listener")
	cli.AddServiceFlags(root, "sessions")
	cli.AddServiceFlags(root, "subscriptions")
	cli.AddServiceFlags(root, "queues")
	cli.AddServiceFlags(root, "messages")
	cli.AddServiceFlags(root, "kv")
	root.Flags().IntP("api-tcp-port", "", 0, "Start API TCP listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("api-tls-port", "", 0, "Start API TLS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("tcp-port", "t", 0, "Start TCP listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("tls-port", "s", 0, "Start TLS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("wss-port", "w", 0, "Start Secure WS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("ws-port", "", 0, "Start WS listener on this port. Specify 0 to disable the listener")
	root.Flags().StringP("tls-cn", "", "localhost", "Get ACME certificat for this CN")
	root.Flags().BoolP("with-router", "", false, "Start the router message consumer")
	root.Execute()
}
