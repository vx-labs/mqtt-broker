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

	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/listener"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"

	mqttConfig "github.com/vx-labs/iot-mqtt-config"
	"github.com/vx-labs/mqtt-broker/network"
	tlsProvider "github.com/vx-labs/mqtt-broker/tls/api"

	"github.com/google/uuid"

	_ "net/http/pprof"

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
			ctx := context.Background()
			nodes, _ := cmd.Flags().GetStringArray("join")
			tcpPort, _ := cmd.Flags().GetInt("tcp-port")
			tlsPort, _ := cmd.Flags().GetInt("tls-port")
			wssPort, _ := cmd.Flags().GetInt("wss-port")
			wsPort, _ := cmd.Flags().GetInt("ws-port")
			clusterNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_CLUSTER)
			serviceNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_SERVICE)
			serviceGossipNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_SERVICE_GOSSIP)
			pprof, _ := cmd.Flags().GetBool("pprof")
			sigc := make(chan os.Signal, 1)

			var tlsConfig *tls.Config
			if pprof {
				go func() {
					log.Printf("INFO: enable pprof endpoint on port 8080")
					log.Println(http.ListenAndServe(":8080", nil))
				}()
			}

			go serveHTTPHealth()
			id := uuid.New().String()

			if os.Getenv("NOMAD_ALLOC_ID") != "" && (tlsPort > 0 || wssPort > 0) {
				consulAPI, vaultAPI, err := mqttConfig.DefaultClients()
				if err != nil {
					panic(err)
				}
				tlsConfig = tlsConfigFromVault(consulAPI, vaultAPI)
				peers, err := ConsulPeers(consulAPI, "cluster", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
				if err == nil {
					nodes = append(nodes, peers...)
				}
			}
			mesh := cluster.New(cluster.Config{
				BindPort:      clusterNetConf.BindPort,
				AdvertisePort: clusterNetConf.AdvertisedPort,
				AdvertiseAddr: clusterNetConf.AdvertisedAddress,
				ID:            id,
			})
			log.Printf(clusterNetConf.Describe(FLAG_NAME_CLUSTER))
			log.Printf("INFO: use the following address to join the cluster: %s:%d", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)

			mesh.Join(nodes)
			lis := listener.New(id, mesh, listener.Config{
				TCPPort: tcpPort,
				TLS:     tlsConfig,
				TLSPort: tlsPort,
				WSPort:  wsPort,
				WSSPort: wssPort,
			})
			addr := listener.Serve(lis, serviceNetConf.BindPort)
			port := addr.Addr().(*net.TCPAddr).Port
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
			cluster.NewServiceLayer("listener", serviceConfig, mesh)

			quit := make(chan struct{})
			signal.Notify(sigc,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)
			log.Printf("INFO: listener service started")
			go func() {
				defer close(quit)
				<-sigc
				log.Printf("INFO: received termination signal")
				log.Printf("INFO: leaving cluster")
				mesh.Leave()
				log.Printf("INFO: left cluster")
				log.Printf("INFO: stopping listener")
				lis.Close(ctx)
				addr.Close()
				log.Printf("INFO: listener stopped")
			}()
			<-quit
		},
	}
	root.Flags().StringArrayP("join", "j", nil, "Join this node")
	root.Flags().BoolP("pprof", "", false, "Enable pprof endpoint")
	root.Flags().IntP("tcp-port", "t", 0, "Start TCP listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("tls-port", "s", 0, "Start TLS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("wss-port", "w", 0, "Start Secure WS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("ws-port", "", 0, "Start WS listener on this port. Specify 0 to disable the listener")
	network.RegisterFlagsForService(root, FLAG_NAME_CLUSTER, 3500)
	network.RegisterFlagsForService(root, FLAG_NAME_SERVICE_GOSSIP, 0)
	network.RegisterFlagsForService(root, FLAG_NAME_SERVICE, 3100)
	root.Execute()
}

func serveHTTPHealth() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	log.Println(http.ListenAndServe("[::]:9000", mux))
}
