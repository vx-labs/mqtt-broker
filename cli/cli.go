package cli

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/network"
)

const (
	FLAG_NAME_CLUSTER        = "cluster"
	FLAG_NAME_SERVICE        = "service"
	FLAG_NAME_SERVICE_GOSSIP = "gossip"
)

type Service interface {
	Serve(port int) net.Listener
	Shutdown()
	JoinServiceLayer(layer cluster.ServiceLayer)
}

func AddClusterFlags(root *cobra.Command) {
	root.Flags().StringSliceP("join", "j", []string{}, "Join this node")
	viper.BindPFlag("join", root.Flags().Lookup("join"))

	root.Flags().BoolP("pprof", "", false, "Enable pprof endpoint")
	viper.BindPFlag("pprof", root.Flags().Lookup("pprof"))
	network.RegisterFlagsForService(root, FLAG_NAME_CLUSTER, 3500)
	network.RegisterFlagsForService(root, FLAG_NAME_SERVICE_GOSSIP, 0)
	network.RegisterFlagsForService(root, FLAG_NAME_SERVICE, 0)
}

func Run(cmd *cobra.Command, name string, serviceFunc func(id string, mesh cluster.Mesh) Service) {
	if viper.GetBool("pprof") {
		go func() {
			log.Printf("INFO: enable pprof endpoint on port 8080")
			log.Println(http.ListenAndServe(":8080", nil))
		}()
	}
	sigc := make(chan os.Signal, 1)

	clusterNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_CLUSTER)
	serviceNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_SERVICE)
	serviceGossipNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_SERVICE_GOSSIP)
	id := uuid.New().String()
	go serveHTTPHealth()
	mesh := joinMesh(id, clusterNetConf)

	service := serviceFunc(id, mesh)
	nodes := viper.GetStringSlice("join")
	mesh.Join(nodes)

	listener := service.Serve(serviceNetConf.BindPort)
	port := listener.Addr().(*net.TCPAddr).Port

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
	layer := cluster.NewServiceLayer(name, serviceConfig, mesh)
	service.JoinServiceLayer(layer)

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
		log.Printf("INFO: cluster left")
		log.Printf(fmt.Sprintf("INFO: stopping %s", name))
		service.Shutdown()
		log.Printf(fmt.Sprintf("INFO: %s stopped", name))
		log.Printf(fmt.Sprintf("INFO: stopping %s RPC listener", name))
		listener.Close()
		log.Printf(fmt.Sprintf("INFO: %s RPC listener stopped", name))
		log.Printf(fmt.Sprintf("INFO: %s stopped", name))
	}()
	<-quit

}

func serveHTTPHealth() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	err := http.ListenAndServe("[::]:9000", mux)
	if err != nil {
		log.Printf("ERROR: failed to run healthcheck endpoint: %v", err)
	}
}

func joinMesh(id string, clusterNetConf network.Configuration) cluster.Mesh {
	mesh := cluster.New(cluster.Config{
		BindPort:      clusterNetConf.BindPort,
		AdvertisePort: clusterNetConf.AdvertisedPort,
		AdvertiseAddr: clusterNetConf.AdvertisedAddress,
		ID:            id,
	})
	log.Printf(clusterNetConf.Describe(FLAG_NAME_CLUSTER))
	log.Printf("INFO: use the following address to join the cluster: %s:%d", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
	return mesh
}
