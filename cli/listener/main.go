package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"log"
	"os"

	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/listener"
	"go.uber.org/zap"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"

	mqttConfig "github.com/vx-labs/iot-mqtt-config"
	tlsProvider "github.com/vx-labs/mqtt-broker/tls/api"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
)

func getTLSProvider(consulAPI *consul.Client, vaultAPI *vault.Client, email string) *tlsProvider.Client {
	opts := []tlsProvider.Opt{
		tlsProvider.WithEmail(email),
	}
	if os.Getenv("LE_STAGING") == "true" {
		opts = append(opts, tlsProvider.WithStagingAPI())
	}
	c, err := tlsProvider.New(
		consulAPI, vaultAPI,
		opts...,
	)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func tlsConfigFromVault(consulAPI *consul.Client, vaultAPI *vault.Client) *tls.Config {
	ctx := context.Background()
	tlsAppConfig, _, err := mqttConfig.TLS(consulAPI)
	if err != nil {
		panic(err)
	}
	cn := tlsAppConfig.CN
	email := tlsAppConfig.LetsEncryptAccountEmail
	api := getTLSProvider(consulAPI, vaultAPI, email)

	log.Printf("INFO: fetching TLS configuration for CN=%s", cn)
	certs, err := api.GetCertificate(ctx, cn)
	if err != nil {
		log.Fatalf("unable to fetch certificate from tls provider: %v", err)
	}
	return &tls.Config{
		Certificates: certs,
		Rand:         rand.Reader,
	}

}

func main() {
	root := &cobra.Command{
		Use: "listener",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Run(cmd, "listener", func(id string, logger *zap.Logger, mesh cluster.Mesh) cli.Service {
				tcpPort, _ := cmd.Flags().GetInt("tcp-port")
				tlsPort, _ := cmd.Flags().GetInt("tls-port")
				wssPort, _ := cmd.Flags().GetInt("wss-port")
				wsPort, _ := cmd.Flags().GetInt("ws-port")
				var tlsConfig *tls.Config

				if os.Getenv("NOMAD_ALLOC_ID") != "" && (tlsPort > 0 || wssPort > 0) {
					consulAPI, vaultAPI, err := mqttConfig.DefaultClients()
					if err != nil {
						panic(err)
					}
					tlsConfig = tlsConfigFromVault(consulAPI, vaultAPI)

				}
				return listener.New(id, logger, mesh, listener.Config{
					TCPPort: tcpPort,
					TLS:     tlsConfig,
					TLSPort: tlsPort,
					WSPort:  wsPort,
					WSSPort: wssPort,
				})
			})
		},
	}
	cli.AddClusterFlags(root)
	root.Flags().IntP("tcp-port", "t", 0, "Start TCP listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("tls-port", "s", 0, "Start TLS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("wss-port", "w", 0, "Start Secure WS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("ws-port", "", 0, "Start WS listener on this port. Specify 0 to disable the listener")
	root.Execute()
}
