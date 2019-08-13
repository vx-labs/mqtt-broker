package cli

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	consul "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/cluster/types"
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
	JoinServiceLayer(layer types.ServiceLayer)
	Health() string
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

func JoinConsulPeers(api *consul.Client, service string, selfAddress string, selfPort int, mesh cluster.Mesh, logger *zap.Logger) error {
	foundSelf := false

	var index uint64
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		services, meta, err := api.Health().Service(
			service,
			"",
			false,
			&consul.QueryOptions{
				WaitIndex: index,
				WaitTime:  15 * time.Second,
			},
		)
		if err != nil {
			<-ticker.C
			continue
		}
		index = meta.LastIndex
		peers := []string{}
		for _, service := range services {
			logger.Info("discovered node", zap.String("node_address", service.Service.Address), zap.Int("node_port", service.Service.Port), zap.String("node_health", service.Checks.AggregatedStatus()))
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
		if foundSelf && len(peers) >= 3 {
			if mesh.Join(peers) == nil {
				return nil
			}
		}
	}
}

func logService(logger *zap.Logger, id, name string, config network.Configuration) {
	logger.Info("loaded service config",
		zap.String("bind_address", config.BindAddress),
		zap.Int("bind_port", config.BindPort),
		zap.String("advertised_address", config.AdvertisedAddress),
		zap.Int("advertised_port", config.AdvertisedPort),
	)
}

func Run(cmd *cobra.Command, name string, serviceFunc func(id string, logger *zap.Logger, mesh cluster.Mesh) Service) {
	id := uuid.New().String()

	if viper.GetBool("pprof") {
		go func() {
			fmt.Println("pprof endpoint is running on port 8080")
			http.ListenAndServe(":8080", nil)
		}()
	}
	sigc := make(chan os.Signal, 1)
	var logger *zap.Logger
	var err error
	fields := []zap.Field{
		zap.String("node_id", id), zap.String("service_name", name),
	}
	if allocID := os.Getenv("NOMAD_ALLOC_ID"); allocID != "" {
		fields = append(fields,
			zap.String("nomad_alloc_id", os.Getenv("NOMAD_ALLOC_ID")),
			zap.String("nomad_alloc_name", os.Getenv("NOMAD_ALLOC_NAME")),
			zap.String("nomad_alloc_index", os.Getenv("NOMAD_ALLOC_INDEX")),
		)
	}

	opts := []zap.Option{
		zap.Fields(fields...),
	}
	if os.Getenv("ENABLE_PRETTY_LOG") == "true" {
		logger, err = zap.NewDevelopment(opts...)
	} else {
		logger, err = zap.NewProduction(opts...)
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	clusterNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_CLUSTER)
	serviceNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_SERVICE)
	serviceGossipNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_SERVICE_GOSSIP)

	mesh := joinMesh(id, logger, clusterNetConf)
	logService(logger, id, FLAG_NAME_CLUSTER, clusterNetConf)

	service := serviceFunc(id, logger, mesh)
	var clusterMemberFound chan struct{}
	if allocID := os.Getenv("NOMAD_ALLOC_ID"); allocID != "" {
		consulConfig := consul.DefaultConfig()
		consulConfig.HttpClient = http.DefaultClient
		consulAPI, err := consul.NewClient(consulConfig)
		if err != nil {
			logger.Fatal("failed to connect to consul")
		}
		clusterMemberFound = make(chan struct{})
		go func() {
			JoinConsulPeers(consulAPI, "cluster", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort, mesh, logger)
			close(clusterMemberFound)
		}()
	}

	go serveHTTPHealth(logger, mesh, service)
	nodes := viper.GetStringSlice("join")
	mesh.Join(nodes)
	if clusterMemberFound != nil {
		<-clusterMemberFound
	}
	listener := service.Serve(serviceNetConf.BindPort)
	if listener != nil {
		port := listener.Addr().(*net.TCPAddr).Port
		logService(logger, id, FLAG_NAME_SERVICE, serviceNetConf)
		logService(logger, id, FLAG_NAME_SERVICE_GOSSIP, serviceGossipNetConf)

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
		layer := cluster.NewServiceLayer(name, logger, serviceConfig, mesh)
		service.JoinServiceLayer(layer)
	}
	quit := make(chan struct{})
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		defer close(quit)
		<-sigc
		logger.Info("received termination signal")
		mesh.Leave()
		logger.Info("cluster left")
		service.Shutdown()
		logger.Info("stopped service")
		if listener != nil {
			listener.Close()
			logger.Info("stopped rpc listener")
		}
	}()
	<-quit
}

type healthChecker interface {
	Health() string
}

func serveHTTPHealth(logger *zap.Logger, mesh healthChecker, service healthChecker) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		for _, status := range []string{mesh.Health(), service.Health()} {
			switch status {
			case "warning":
				w.WriteHeader(http.StatusTooManyRequests)
				return
			case "critical":
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	err := http.ListenAndServe("[::]:9000", mux)
	if err != nil {
		logger.Error("failed to run healthcheck endpoint", zap.Error(err))
	}
}

func joinMesh(id string, logger *zap.Logger, clusterNetConf network.Configuration) cluster.Mesh {
	mesh := cluster.New(logger, cluster.Config{
		BindPort:      clusterNetConf.BindPort,
		AdvertisePort: clusterNetConf.AdvertisedPort,
		AdvertiseAddr: clusterNetConf.AdvertisedAddress,
		ID:            id,
	})
	fmt.Printf("Use the following address to join the cluster: %s:%d\n", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
	return mesh
}
