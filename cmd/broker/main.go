package main

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vx-labs/mqtt-broker/listener"

	"google.golang.org/grpc"

	"github.com/google/uuid"

	"github.com/vx-labs/mqtt-broker/transport"

	_ "net/http/pprof"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/vx-labs/iot-mqtt-auth/api"
	mqttConfig "github.com/vx-labs/iot-mqtt-config"
	tlsProvider "github.com/vx-labs/mqtt-broker/tls/api"

	"github.com/spf13/cobra"

	"github.com/vx-labs/mqtt-broker/broker"

	"github.com/vx-labs/mqtt-broker/identity"
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
func ConsulPeers(api *consul.Client, service string, self identity.Identity) ([]string, error) {
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
			if service.Service.Address == self.Public().Host() &&
				service.Service.Port == self.Public().Port() {
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
func makeSessionID(tenant string, clientID []byte) (string, error) {
	hash := sha1.New()
	_, err := hash.Write([]byte(tenant))
	if err != nil {
		return "", err
	}
	_, err = hash.Write(clientID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
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
func main() {
	root := cobra.Command{
		Use: "broker",
		Run: func(cmd *cobra.Command, args []string) {
			nodes, _ := cmd.Flags().GetStringArray("join")
			tcpPort, _ := cmd.Flags().GetInt("tcp-port")
			tlsPort, _ := cmd.Flags().GetInt("tls-port")
			wssPort, _ := cmd.Flags().GetInt("wss-port")
			wsPort, _ := cmd.Flags().GetInt("ws-port")
			rpcPort, _ := cmd.Flags().GetInt("rpc-port")
			gossipPort, _ := cmd.Flags().GetInt("gossip-port")
			nomad, _ := cmd.Flags().GetBool("nomad")
			pprof, _ := cmd.Flags().GetBool("pprof")
			useVault, _ := cmd.Flags().GetBool("use-vault")
			useConsul, _ := cmd.Flags().GetBool("use-consul")
			useVXAuth, _ := cmd.Flags().GetBool("use-vx-auth")
			natsURL, _ := cmd.Flags().GetString("nats-streaming-url")
			sigc := make(chan os.Signal, 1)

			var id identity.Identity
			var rpcId identity.Identity
			var err error
			var tlsConfig *tls.Config
			var consulAPI *consul.Client
			var vaultAPI *vault.Client
			if pprof {
				go func() {
					log.Printf("INFO: enable pprof endpoint on port 8080")
					log.Println(http.ListenAndServe(":8080", nil))
				}()
			}

			if nomad {
				id, err = identity.NomadService("broker")
				rpcId, err = identity.NomadService("rpc")
			} else if gossipPort > 0 {
				id = identity.StaticService(gossipPort)
			} else {
				id, err = identity.LocalService()
			}
			if err != nil {
				log.Fatalf("FATAL: unable to determine service host and ports.")
			}
			go serveHTTPHealth()
			config := broker.DefaultConfig()
			config.TCPPort = tcpPort
			config.TLSPort = tlsPort
			config.WSSPort = wssPort
			config.RPCPort = rpcPort
			config.WSPort = wsPort
			config.NATSURL = natsURL
			if rpcId != nil {
				config.RPCIdentity = rpcId
			}

			if useVault || useConsul {
				consulAPI, vaultAPI, err = mqttConfig.DefaultClients()
				if err != nil {
					panic(err)
				}
				if useVault {
					tlsConfig = tlsConfigFromVault(consulAPI, vaultAPI)
				}
			}
			if useVXAuth {
				config.AuthHelper = authHelper(context.Background())
			}
			config.TLS = tlsConfig
			id = id.WithID(uuid.New().String())
			listenerConn, err := grpc.Dial("localhost:3001", grpc.WithInsecure())
			if err != nil {
				panic(err)
			}
			listenerClient := listener.NewClient(listenerConn)

			brokerConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", config.RPCPort), grpc.WithInsecure())
			if err != nil {
				panic(err)
			}
			brokerClient := broker.NewClient(brokerConn)

			lis := listener.New(brokerClient, listener.Config{
				TCPPort: config.TCPPort,
				TLS:     config.TLS,
				TLSPort: config.TLSPort,
				WSPort:  config.WSPort,
				WSSPort: config.WSSPort,
			})
			go listener.Serve(lis, 3001)
			instance := broker.New(id, listenerClient, config)
			if useConsul {
				go func() {
					nodes, err = ConsulPeers(consulAPI, "broker", id)
					log.Printf("INFO: started broker instance %s on %s", id.ID(), id.Public().String())
					if len(nodes) > 0 {
						instance.Join(nodes)
					}
				}()
			}

			quit := make(chan struct{})
			signal.Notify(sigc,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)
			go func() {
				defer close(quit)
				<-sigc
				log.Printf("INFO: received termination signal")
				log.Printf("INFO: stopping broker")
				instance.Stop()
				log.Printf("INFO: broker stopped")
			}()
			<-quit
		},
	}
	root.Flags().StringArrayP("join", "j", nil, "Join this node")
	root.Flags().BoolP("pprof", "", false, "Enable pprof endpoint")
	root.Flags().BoolP("nomad", "", false, "Discover this node identity using Nomad environment variables")
	root.Flags().BoolP("use-vault", "", false, "Manage node certificates and private keys using Vault and Consul")
	root.Flags().BoolP("use-consul", "", false, "Discover other peers using Consul")
	root.Flags().BoolP("use-vx-auth", "", false, "Use VX Authentication Service to authenticate clients")
	root.Flags().IntP("tcp-port", "t", 0, "Start TCP listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("tls-port", "s", 0, "Start TLS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("wss-port", "w", 0, "Start Secure WS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("ws-port", "", 0, "Start WS listener on this port. Specify 0 to disable the listener")
	root.Flags().IntP("rpc-port", "r", 3000, "Start GRPC listener on this port.")
	root.Flags().IntP("gossip-port", "g", 0, "Use this port for Mesh traffic. Specify 0 to use a random port")
	root.Flags().StringP("nats-streaming-url", "", "", "Export published message to this NATS-Streaming service")
	root.Execute()
}

func serveHTTPHealth() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	log.Println(http.ListenAndServe("[::]:9000", mux))
}
