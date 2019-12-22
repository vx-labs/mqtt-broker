package cli

import (
	"encoding/json"
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
	clusterconfig "github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/network"
)

const (
	FLAG_NAME_CLUSTER            = "cluster"
	FLAG_NAME_SERVICE            = "service"
	FLAG_NAME_SERVICE_GOSSIP     = "gossip"
	FLAG_NAME_SERVICE_GOSSIP_RPC = "gossip_rpc"
)

type Service interface {
	Serve(port int) net.Listener
	Shutdown()
	JoinServiceLayer(string, *zap.Logger, cluster.ServiceConfig, cluster.ServiceConfig, cluster.DiscoveryLayer)
	Health() string
}

func AddClusterFlags(root *cobra.Command) {
	root.Flags().StringSliceP("join", "j", []string{}, "Join this node")
	viper.BindPFlag("join", root.Flags().Lookup("join"))

	root.Flags().IntP("healthcheck-port", "", 9000, "Run healthcheck http server on this port")
	viper.BindPFlag("healthcheck-port", root.Flags().Lookup("healthcheck-port"))

	root.Flags().BoolP("pprof", "", false, "Enable pprof endpoint")
	root.Flags().BoolP("debug", "", false, "Enable debug logs and fancy log printing")
	viper.BindPFlag("pprof", root.Flags().Lookup("pprof"))
	viper.BindPFlag("debug", root.Flags().Lookup("debug"))
	network.RegisterFlagsForService(root, FLAG_NAME_CLUSTER, 3500)
}
func AddServiceFlags(root *cobra.Command, name string) {
	network.RegisterFlagsForService(root, fmt.Sprintf("%s_gossip_rpc", name), 0)
	network.RegisterFlagsForService(root, fmt.Sprintf("%s_gossip", name), 0)
	network.RegisterFlagsForService(root, name, 0)
}

func JoinConsulPeers(api *consul.Client, service string, selfAddress string, selfPort int, mesh cluster.Mesh, logger *zap.Logger) error {
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
			if service.Checks.AggregatedStatus() == consul.HealthCritical {
				continue
			}
			if service.Service.Address == selfAddress &&
				service.Service.Port == selfPort {
				continue
			}
			peer := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)
			peers = append(peers, peer)
		}
		if len(peers) > 0 {
			mesh.Join(peers)
		}
	}
}

func logService(logger *zap.Logger, name string, config network.Configuration) {
	fmt.Printf("service %s is listening on %s:%d\n", name, config.AdvertisedAddress, config.AdvertisedPort)
	logger.Debug("loaded service config",
		zap.String("bind_address", config.BindAddress),
		zap.Int("bind_port", config.BindPort),
		zap.String("advertised_address", config.AdvertisedAddress),
		zap.Int("advertised_port", config.AdvertisedPort),
	)
}

type serviceRunConfig struct {
	Gossip    network.Configuration
	GossipRPC network.Configuration
	Network   network.Configuration
	Service   Service
	ID        string
}

func (s *serviceRunConfig) Health() string {
	return s.Service.Health()
}
func (s *serviceRunConfig) ServiceName() string {
	return s.ID
}

type Context struct {
	ID          string
	Logger      *zap.Logger
	Discovery   cluster.DiscoveryLayer
	MeshNetConf *network.Configuration
	Services    []*serviceRunConfig
}

func (ctx *Context) AddService(cmd *cobra.Command, name string, f func(id string, logger *zap.Logger, mesh cluster.DiscoveryLayer) Service) {
	id := ctx.ID
	logger := ctx.Logger.WithOptions(zap.Fields(zap.String("service_name", name)))
	serviceNetConf := network.ConfigurationFromFlags(cmd, name)
	serviceGossipNetConf := network.ConfigurationFromFlags(cmd, fmt.Sprintf("%s_gossip", name))
	serviceGossipRPCNetConf := network.ConfigurationFromFlags(cmd, fmt.Sprintf("%s_gossip_rpc", name))
	service := f(id, logger, ctx.Discovery)

	ctx.Services = append(ctx.Services, &serviceRunConfig{
		Gossip:    serviceGossipNetConf,
		GossipRPC: serviceGossipRPCNetConf,
		Network:   serviceNetConf,
		Service:   service,
		ID:        name,
	})
}

func (ctx *Context) Run() error {
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error("panic", zap.String("panic_log", fmt.Sprint(r)))
		}
		ctx.Logger.Sync()
		os.Exit(9)
	}()
	logger := ctx.Logger
	clusterNetConf := ctx.MeshNetConf
	mesh := ctx.Discovery

	if allocID := os.Getenv("NOMAD_ALLOC_ID"); allocID != "" {
		logger.Debug("nomad environment detected, attempting to find peers using Consul discovery API")
		consulConfig := consul.DefaultConfig()
		consulConfig.HttpClient = http.DefaultClient
		consulAPI, err := consul.NewClient(consulConfig)
		if err != nil {
			logger.Fatal("failed to connect to consul")
		}
		go JoinConsulPeers(consulAPI, "cluster", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort, mesh, logger)
	}
	sensors := make([]healthChecker, 0, len(ctx.Services)+1)
	sensors = append(sensors, mesh)
	for _, service := range ctx.Services {
		sensors = append(sensors, service)
	}
	go serveHTTPHealth(viper.GetInt("healthcheck-port"), logger, sensors)
	nodes := viper.GetStringSlice("join")
	if len(nodes) > 0 {
		go func() {
			joinTicker := time.NewTicker(5 * time.Second)
			defer joinTicker.Stop()
			retries := 5
			for retries > 0 {
				err := mesh.Join(nodes)
				if err == nil {
					return
				}
				logger.Warn("failed to join provided cluster node", zap.Error(err))
				retries--
				<-joinTicker.C
			}
		}()
	}
	for _, service := range ctx.Services {
		logger := ctx.Logger.WithOptions(zap.Fields(zap.String("service_name", service.ID)))
		serviceNetConf := service.Network
		serviceGossipNetConf := service.Gossip
		serviceGossipRPCNetConf := service.GossipRPC

		listener := service.Service.Serve(serviceNetConf.BindPort)
		if listener != nil {
			port := listener.Addr().(*net.TCPAddr).Port

			serviceConfig := cluster.ServiceConfig{
				AdvertiseAddr: serviceGossipNetConf.AdvertisedAddress,
				AdvertisePort: serviceGossipNetConf.AdvertisedPort,
				BindPort:      serviceGossipNetConf.BindPort,
				ID:            ctx.ID,
				ServicePort:   serviceNetConf.AdvertisedPort,
			}
			if serviceConfig.AdvertisePort == 0 {
				serviceConfig.AdvertisePort = serviceConfig.BindPort
				serviceConfig.ServicePort = port
			}
			gossipRPCConfig := cluster.ServiceConfig{
				AdvertiseAddr: serviceGossipRPCNetConf.AdvertisedAddress,
				AdvertisePort: serviceGossipRPCNetConf.AdvertisedPort,
				BindPort:      serviceGossipRPCNetConf.BindPort,
				ID:            ctx.ID,
				ServicePort:   serviceGossipRPCNetConf.AdvertisedPort,
			}
			if gossipRPCConfig.AdvertisePort == 0 {
				gossipRPCConfig.AdvertisePort = gossipRPCConfig.BindPort
				gossipRPCConfig.ServicePort = port
			}
			logService(logger, service.ID, service.Network)
			service.Service.JoinServiceLayer(service.ID, logger, serviceConfig, gossipRPCConfig, ctx.Discovery)
		}
	}
	quit := make(chan struct{})
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		defer close(quit)
		<-sigc
		logger.Info("received termination signal")
		for _, service := range ctx.Services {
			logger.Debug(fmt.Sprintf("stopping service %s", service.ID))
			service.Service.Shutdown()
			logger.Debug(fmt.Sprintf("stopped service %s", service.ID))
		}
		ctx.Discovery.Leave()
		logger.Info("cluster left")
	}()
	<-quit
	return nil
}

func Bootstrap(cmd *cobra.Command) *Context {
	id := uuid.New().String()
	ctx := &Context{
		ID: id,
	}
	var logger *zap.Logger
	var err error
	fields := []zap.Field{
		zap.String("node_id", id[:8]), zap.String("version", Version()),
	}
	if allocID := os.Getenv("NOMAD_ALLOC_ID"); allocID != "" {
		fields = append(fields,
			zap.String("nomad_alloc_id", os.Getenv("NOMAD_ALLOC_ID")[:8]),
			zap.String("nomad_alloc_name", os.Getenv("NOMAD_ALLOC_NAME")),
			zap.String("nomad_alloc_index", os.Getenv("NOMAD_ALLOC_INDEX")),
		)
	}
	opts := []zap.Option{
		zap.Fields(fields...),
	}
	if viper.GetBool("debug") {
		logger, err = zap.NewDevelopment(opts...)
		logger.Debug("started debug logger")
	} else {
		logger, err = zap.NewProduction(opts...)
	}
	if err != nil {
		panic(err)
	}
	ctx.Logger = logger
	if viper.GetBool("pprof") {
		go func() {
			fmt.Println("pprof endpoint is running on port 8080")
			http.ListenAndServe(":8080", nil)
		}()
	}
	clusterNetConf := network.ConfigurationFromFlags(cmd, FLAG_NAME_CLUSTER)
	ctx.MeshNetConf = &clusterNetConf
	mesh := createMesh(id, logger, clusterNetConf)
	ctx.Discovery = mesh
	return ctx
}

type healthChecker interface {
	Health() string
	ServiceName() string
}
type ServiceHealthReport struct {
	ServiceName string `json:"service_name"`
	Health      string `json:"health"`
}
type HealthReport struct {
	Timestamp    time.Time             `json:"timestamp"`
	GlobalHealth string                `json:"global_health"`
	Services     []ServiceHealthReport `json:"services"`
}

var serviceHealthMap = map[string]int{
	"ok":       0,
	"warning":  1,
	"critical": 2,
}

func serveHTTPHealth(port int, logger *zap.Logger, sensors []healthChecker) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		report := HealthReport{
			Timestamp: time.Now(),
		}
		worst := "ok"
		for _, sensor := range sensors {
			status := sensor.Health()
			report.Services = append(report.Services, ServiceHealthReport{
				Health:      status,
				ServiceName: sensor.ServiceName(),
			})
			if serviceHealthMap[status] > serviceHealthMap[worst] {
				worst = status
			}
		}
		report.GlobalHealth = worst
		switch worst {
		case "warning":
			w.WriteHeader(http.StatusTooManyRequests)
		case "critical":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(report)
	})
	err := http.ListenAndServe(fmt.Sprintf("[::]:%d", port), mux)
	if err != nil {
		logger.Warn("failed to run healthcheck endpoint", zap.Error(err))
	}
}

func createMesh(id string, logger *zap.Logger, clusterNetConf network.Configuration) cluster.DiscoveryLayer {
	mesh := cluster.NewDiscoveryLayer(logger, clusterconfig.Config{
		BindPort:      clusterNetConf.BindPort,
		AdvertisePort: clusterNetConf.AdvertisedPort,
		AdvertiseAddr: clusterNetConf.AdvertisedAddress,
		ID:            id,
	})
	fmt.Printf("Use the following address to join the cluster: %s:%d\n", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
	return mesh
}
