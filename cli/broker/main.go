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
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/transport"

	"github.com/spf13/viper"

	"github.com/vx-labs/mqtt-broker/broker"

	"github.com/vx-labs/mqtt-broker/cluster"

	_ "net/http/pprof"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/vx-labs/iot-mqtt-auth/api"

	mqttConfig "github.com/vx-labs/iot-mqtt-config"
	tlsProvider "github.com/vx-labs/mqtt-broker/tls/api"

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
		log.Println("INFO: calling VX auth handler")
		defer func() {
			log.Println("INFO: VX auth handler returned")
		}()
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
		Use: "broker",
		Run: func(cmd *cobra.Command, args []string) {
			clusterNetConf := network.ConfigurationFromFlags(cmd, cli.FLAG_NAME_CLUSTER)

			cli.Run(cmd, "broker", func(id string, mesh cluster.Mesh) cli.Service {
				config := broker.DefaultConfig()
				if os.Getenv("NOMAD_ALLOC_ID") != "" {
					nodes := viper.GetStringSlice("join")
					consulAPI, _, err := mqttConfig.DefaultClients()
					if err != nil {
						panic(err)
					}
					peers, err := ConsulPeers(consulAPI, "cluster", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
					if err == nil {
						nodes = append(nodes, peers...)
						viper.Set("join", nodes)
					}
					config.AuthHelper = authHelper(context.Background())
				}
				return broker.New(id, mesh, config)
			})
		},
	}
	cli.AddClusterFlags(root)
	root.Execute()
}
