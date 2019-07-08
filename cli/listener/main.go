package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/listener"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"

	mqttConfig "github.com/vx-labs/iot-mqtt-config"
	"github.com/vx-labs/mqtt-broker/network"
	tlsProvider "github.com/vx-labs/mqtt-broker/tls/api"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func ConsulPeers(api *consul.Client, service string, selfAddress string, selfPort int) ([]string, error) {
	foundSelf := false
	var (
		services []*consul.ServiceEntry
		err      error
	)
	opts := &consul.QueryOptions{}
	for {
		services, _, err = api.Health().Service(
			service,
			"",
			true,
			opts,
		)
		if err != nil {
			return nil, err
		}
		peers := []string{}
		for _, service := range services {
			if service.Checks.AggregatedStatus() == consul.HealthCritical {
				continue
			}
			if service.Service.Address == selfAddress &&
				service.Service.Port == selfPort {
				foundSelf = true
				continue
			}
			peer := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)
			peers = append(peers, peer)
		}
		if foundSelf && len(peers) > 0 {
			return peers, nil
		}
		log.Println("INFO: waiting for other peers to appear on consul registry")
		time.Sleep(3 * time.Second)
	}
}

func main() {
	root := &cobra.Command{
		Use: "listener",
		Run: func(cmd *cobra.Command, args []string) {

			clusterNetConf := network.ConfigurationFromFlags(cmd, cli.FLAG_NAME_CLUSTER)

			cli.Run(cmd, "listener", func(id string, mesh cluster.Mesh) cli.Service {
				tcpPort, _ := cmd.Flags().GetInt("tcp-port")
				tlsPort, _ := cmd.Flags().GetInt("tls-port")
				wssPort, _ := cmd.Flags().GetInt("wss-port")
				wsPort, _ := cmd.Flags().GetInt("ws-port")
				var tlsConfig *tls.Config

				if os.Getenv("NOMAD_ALLOC_ID") != "" && (tlsPort > 0 || wssPort > 0) {
					nodes := viper.GetStringSlice("join")
					consulAPI, vaultAPI, err := mqttConfig.DefaultClients()
					if err != nil {
						panic(err)
					}
					tlsConfig = tlsConfigFromVault(consulAPI, vaultAPI)
					peers, err := ConsulPeers(consulAPI, "cluster", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
					if err == nil {
						nodes = append(nodes, peers...)
						viper.Set("join", nodes)
					}
				}
				return listener.New(id, mesh, listener.Config{
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
