package cli

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
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
	JoinServiceLayer(string, *zap.Logger, cluster.ServiceConfig, cluster.ServiceConfig, discovery.DiscoveryAdapter)
	Health() string
}

func AddClusterFlags(root *cobra.Command, config *viper.Viper) {
	root.Flags().StringP("node-id", "", uuid.New().String(), "This node cluster-wide unique ID")
	config.BindPFlag("node-id", root.Flags().Lookup("node-id"))
	config.BindEnv("node-id", "NOMAD_ALLOC_ID")

	root.Flags().StringSliceP("join", "j", []string{}, "Join this node")
	config.BindPFlag("join", root.Flags().Lookup("join"))
	root.Flags().StringP("grpc-tls-certificate", "", "./run_config/cert.pem", "TLS certificate to use")
	config.BindPFlag("grpc-tls-certificate", root.Flags().Lookup("grpc-tls-certificate"))
	root.Flags().StringP("grpc-tls-private-key", "", "./run_config/privkey.pem", "TLS private key to use")
	config.BindPFlag("grpc-tls-private-key", root.Flags().Lookup("grpc-tls-private-key"))
	root.Flags().StringP("grpc-tls-certificate-authority", "", "./run_config/cert.pem", "TLS CA to use")
	config.BindPFlag("grpc-tls-certificate-authority", root.Flags().Lookup("grpc-tls-certificate-authority"))

	root.Flags().IntP("healthcheck-port", "", 9000, "Run healthcheck http server on this port")
	config.BindPFlag("healthcheck-port", root.Flags().Lookup("healthcheck-port"))

	root.Flags().BoolP("pprof", "", false, "Enable pprof endpoint")
	root.Flags().BoolP("debug", "", false, "Enable debug logs and fancy log printing")
	config.BindPFlag("pprof", root.Flags().Lookup("pprof"))
	config.BindPFlag("debug", root.Flags().Lookup("debug"))
	network.RegisterFlagsForService(root, config, FLAG_NAME_CLUSTER, 3500)
}
func AddServiceFlags(root *cobra.Command, config *viper.Viper, name string) {
	startServiceFlagName := fmt.Sprintf("start-%s", name)
	root.Flags().BoolP(startServiceFlagName, "", true, fmt.Sprintf("Start the %s service", name))
	config.BindPFlag(startServiceFlagName, root.Flags().Lookup(startServiceFlagName))

	network.RegisterFlagsForService(root, config, fmt.Sprintf("%s_gossip_rpc", name), 0)
	network.RegisterFlagsForService(root, config, fmt.Sprintf("%s_gossip", name), 0)
	network.RegisterFlagsForService(root, config, name, 0)
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
	Discovery   discovery.DiscoveryAdapter
	MeshNetConf *network.Configuration
	Services    []*serviceRunConfig
}

func (ctx *Context) AddService(cmd *cobra.Command, config *viper.Viper, name string, f func(id string, config *viper.Viper, logger *zap.Logger, mesh discovery.DiscoveryAdapter) Service) {
	id := ctx.ID
	logger := ctx.Logger.WithOptions(zap.Fields(zap.String("service_name", name)))
	serviceNetConf := network.ConfigurationFromFlags(cmd, config, name)
	serviceGossipNetConf := network.ConfigurationFromFlags(cmd, config, fmt.Sprintf("%s_gossip", name))
	serviceGossipRPCNetConf := network.ConfigurationFromFlags(cmd, config, fmt.Sprintf("%s_gossip_rpc", name))
	service := f(id, config, logger, ctx.Discovery)

	ctx.Services = append(ctx.Services, &serviceRunConfig{
		Gossip:    serviceGossipNetConf,
		GossipRPC: serviceGossipRPCNetConf,
		Network:   serviceNetConf,
		Service:   service,
		ID:        name,
	})
}

func (ctx *Context) Run(v *viper.Viper) error {
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error("panic", zap.String("panic_log", fmt.Sprint(r)))
			os.Exit(9)
		}
		ctx.Logger.Sync()
	}()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	logger := ctx.Logger

	sensors := make([]healthChecker, 0, len(ctx.Services)+1)
	for _, service := range ctx.Services {
		sensors = append(sensors, service)
	}
	go serveHTTPHealth(v.GetInt("healthcheck-port"), logger, sensors)

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
	go func() {
		defer close(quit)
		<-sigc
		logger.Info("received termination signal")
		for _, service := range ctx.Services {
			logger.Debug(fmt.Sprintf("stopping service %s", service.ID))
			service.Service.Shutdown()
			logger.Debug(fmt.Sprintf("stopped service %s", service.ID))
		}
		ctx.Discovery.Shutdown()
		logger.Info("cluster left")
	}()
	<-quit
	return nil
}
func makeNodeID(id string) string {
	if len(id) > 8 {
		return id
	}
	hash := sha1.New()
	_, err := hash.Write([]byte(id))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Bootstrap(cmd *cobra.Command, v *viper.Viper) *Context {
	nodeID := v.GetString("node-id")

	ctx := &Context{
		ID: makeNodeID(nodeID),
	}
	var logger *zap.Logger
	var err error
	fields := []zap.Field{
		zap.String("node_id", ctx.ID[:8]), zap.String("version", Version()),
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
	if v.GetBool("debug") {
		logger, err = zap.NewDevelopment(opts...)
		logger.Debug("started debug logger")
	} else {
		logger, err = zap.NewProduction(opts...)
	}
	if err != nil {
		panic(err)
	}
	logger.Info("starting node")
	ctx.Logger = logger
	if v.GetBool("pprof") {
		go func() {
			fmt.Println("pprof endpoint is running on port 8080")
			http.ListenAndServe(":8080", nil)
		}()
	}
	clusterNetConf := network.ConfigurationFromFlags(cmd, v, FLAG_NAME_CLUSTER)
	ctx.MeshNetConf = &clusterNetConf
	mesh := createMesh(ctx.ID, v, logger, clusterNetConf)
	ctx.Discovery = mesh
	logger.Info("mesh created")
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
		logger.Debug("failed to run healthcheck endpoint", zap.Error(err))
	}
}

func createMesh(id string, v *viper.Viper, logger *zap.Logger, clusterNetConf network.Configuration) discovery.DiscoveryAdapter {
	mesh := discovery.Mesh(logger, clusterconfig.Config{
		BindPort:      clusterNetConf.BindPort,
		AdvertisePort: clusterNetConf.AdvertisedPort,
		AdvertiseAddr: clusterNetConf.AdvertisedAddress,
		ID:            id,
		JoinList:      v.GetStringSlice("join"),
	})
	fmt.Printf("Use the following address to join the cluster: %s:%d\n", clusterNetConf.AdvertisedAddress, clusterNetConf.AdvertisedPort)
	return mesh
}
