package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/transport"

	"github.com/spf13/viper"

	"github.com/vx-labs/mqtt-broker/broker"

	"github.com/vx-labs/mqtt-broker/cluster"

	"github.com/google/uuid"

	_ "net/http/pprof"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/vx-labs/iot-mqtt-auth/api"

	mqttConfig "github.com/vx-labs/iot-mqtt-config"
	tlsProvider "github.com/vx-labs/mqtt-broker/tls/api"

	"github.com/spf13/cobra"
)

const (
	FLAG_NAME_CLUSTER        = "cluster"
	FLAG_NAME_SERVICE        = "service"
	FLAG_NAME_SERVICE_GOSSIP = "gossip"
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

	log.Printf("fetching TLS configuration for CN=%s", cn)
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
			nodes := viper.GetStringSlice("join")
			pprof := viper.GetBool("pprof")
			clusterNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_CLUSTER)
			serviceNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_SERVICE)
			serviceGossipNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_SERVICE_GOSSIP)
			sigc := make(chan os.Signal, 1)

			if pprof {
				go func() {
					log.Printf("INFO: enable pprof endpoint on port 8080")
					log.Println(http.ListenAndServe(":8080", nil))
				}()
			}

			go serveHTTPHealth()
			id := uuid.New().String()
			config := broker.DefaultConfig()
			mesh := cluster.New(cluster.Config{
				BindPort:      clusterNetConf.BindPort,
				AdvertisePort: clusterNetConf.AdvertisedPort,
				AdvertiseAddr: clusterNetConf.AdvertisedAddress,
				ID:            id,
			})
			log.Printf(clusterNetConf.Describe(FLAG_NAME_CLUSTER))
			log.Printf("INFO: use the following address to join the cluster: %s:%d", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)

			if os.Getenv("NOMAD_ALLOC_ID") != "" {
				consulAPI, _, err := mqttConfig.DefaultClients()
				if err != nil {
					panic(err)
				}
				peers, err := ConsulPeers(consulAPI, "cluster", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
				if err == nil {
					nodes = append(nodes, peers...)
				}
				config.AuthHelper = authHelper(context.Background())
			}
			mesh.Join(nodes)
			instance := broker.New(id, mesh, config)
			addr := broker.Serve(serviceNetConf.BindPort, instance)
			port := addr.Addr().(*net.TCPAddr).Port
			log.Printf("service broker is listening on RPC port %d", port)
			log.Printf(serviceNetConf.Describe(FLAG_NAME_SERVICE))
			log.Printf(serviceGossipNetConf.Describe(FLAG_NAME_SERVICE_GOSSIP))

			serviceConfig := cluster.ServiceConfig{
				AdvertiseAddr: serviceGossipNetConf.AdvertisedAddress,
				AdvertisePort: serviceGossipNetConf.AdvertisedPort,
				BindPort:      serviceGossipNetConf.BindPort,
				ID:            id,
				ServicePort:   serviceNetConf.AdvertisedPort,
			}
			if serviceConfig.AdvertisePort == 0 {
				serviceConfig.AdvertisePort = serviceConfig.BindPort
				serviceConfig.ServicePort = port
			}
			layer := cluster.NewServiceLayer("broker", serviceConfig, mesh)
			instance.Start(layer)
			quit := make(chan struct{})
			signal.Notify(sigc,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)
			go func() {
				defer close(quit)
				<-sigc
				log.Printf("INFO: received termination signal")
				log.Printf("INFO: leaving cluster")
				mesh.Leave()
				log.Printf("INFO: left cluster")
				log.Printf("INFO: stopping broker")
				instance.Stop()
				log.Printf("INFO: listener broker")
				log.Printf("INFO: stopping broker RPC listener")
				addr.Close()
				log.Printf("INFO: stopped broker RPC listener")
				log.Printf("INFO: broker stopped")
			}()
			<-quit
		},
	}
	root.Flags().StringSliceP("join", "j", []string{}, "Join this node")
	viper.BindPFlag("join", root.Flags().Lookup("join"))

	root.Flags().BoolP("pprof", "", false, "Enable pprof endpoint")
	viper.BindPFlag("pprof", root.Flags().Lookup("pprof"))

	network.RegisterFlagsForService(root, FLAG_NAME_CLUSTER, 3500)
	network.RegisterFlagsForService(root, FLAG_NAME_SERVICE_GOSSIP, 0)
	network.RegisterFlagsForService(root, FLAG_NAME_SERVICE, 3000)

	root.Execute()
}

func serveHTTPHealth() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	log.Println(http.ListenAndServe("[::]:9000", mux))
}
