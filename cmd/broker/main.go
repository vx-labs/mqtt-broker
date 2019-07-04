package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vx-labs/mqtt-broker/network"

	"github.com/spf13/viper"

	"github.com/vx-labs/mqtt-broker/broker"

	"github.com/vx-labs/mqtt-broker/cluster"

	"github.com/google/uuid"

	_ "net/http/pprof"

	"github.com/spf13/cobra"
)

const (
	FLAG_NAME_CLUSTER        = "cluster"
	FLAG_NAME_SERVICE        = "service"
	FLAG_NAME_SERVICE_GOSSIP = "gossip"
)

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
			log.Printf("cluster is listening on port [%d -> %d]", clusterNetConf.AdvertisedPort, clusterNetConf.BindPort)
			log.Printf("use the following address to join the cluster: %s:%d", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
			mesh.Join(nodes)
			instance := broker.New(id, mesh, config)
			addr := broker.Serve(serviceNetConf.BindPort, instance)
			port := addr.Addr().(*net.TCPAddr).Port
			log.Printf("service broker is listening on RPC port %d", port)

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
			log.Printf("service broker is listening on Gossip port [%d -> %d]", serviceConfig.AdvertisePort, serviceConfig.BindPort)
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
