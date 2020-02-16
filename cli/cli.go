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
	"github.com/vx-labs/mqtt-broker/adapters/identity"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/path"
)

const (
	FLAG_NAME_SERVICE            = "rpc"
	FLAG_NAME_SERVICE_GOSSIP     = "cluster"
	FLAG_NAME_SERVICE_GOSSIP_RPC = "cluster_rpc"
)

type Service interface {
	Serve(port int) net.Listener
	Shutdown()
	Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error
	Health() string
}

func AddClusterFlags(root *cobra.Command, config *viper.Viper) {
	root.Flags().StringP("node-id", "", uuid.New().String(), "This node cluster-wide unique ID")
	config.BindPFlag("node-id", root.Flags().Lookup("node-id"))
	config.BindEnv("node-id", "NOMAD_ALLOC_ID")

	root.Flags().StringP("discovery-provider", "", "nomad", "Discovery provider to use")
	config.BindPFlag("discovery-provider", root.Flags().Lookup("discovery-provider"))

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
}
func AddServiceFlags(root *cobra.Command, config *viper.Viper, name string) {
	startServiceFlagName := fmt.Sprintf("start-%s", name)
	root.Flags().BoolP(startServiceFlagName, "", true, fmt.Sprintf("Start the %s service", name))
	config.BindPFlag(startServiceFlagName, root.Flags().Lookup(startServiceFlagName))

	network.RegisterFlagsForService(root, config, fmt.Sprintf("%s-cluster_rpc", name), 0)
	network.RegisterFlagsForService(root, config, fmt.Sprintf("%s-cluster", name), 0)
	network.RegisterFlagsForService(root, config, name, 0)
}

func logService(logger *zap.Logger, name string, config identity.Identity) {
	fmt.Printf("service %s is listening on %s:%d\n", name, config.AdvertisedAddress(), config.AdvertisedPort())
	logger.Debug("loaded service config",
		zap.String("bind_address", config.BindAddress()),
		zap.Int("bind_port", config.BindPort()),
		zap.String("advertised_address", config.AdvertisedAddress()),
		zap.Int("advertised_port", config.AdvertisedPort()),
	)
}

type serviceRunConfig struct {
	service Service
	Name    string `json:"name"`
	ID      string `json:"id"`
}

func (s *serviceRunConfig) Health() string {
	return s.service.Health()
}
func (s *serviceRunConfig) ServiceName() string {
	return s.Name
}
func (s *serviceRunConfig) ServiceID() string {
	return s.ID
}

type Context struct {
	ID              string
	Logger          *zap.Logger
	Discovery       discovery.DiscoveryAdapter
	Services        []*serviceRunConfig
	identityCatalog identity.Catalog
}

func (ctx *Context) AddService(cmd *cobra.Command, config *viper.Viper, name string, f func(id string, config *viper.Viper, logger *zap.Logger, mesh discovery.DiscoveryAdapter) Service) {
	exists := false
	for _, service := range ctx.Services {
		for _, wantedName := range []string{
			name,
			fmt.Sprintf("%s-cluster", name),
			fmt.Sprintf("%s-cluster_rpc", name),
		} {
			exists = true
			if service.Name == wantedName && service.ID != "" {
				config.Set(fmt.Sprintf("%s-service-id", wantedName), service.ID)
			}
		}
	}
	serviceNetConf := network.ConfigurationFromFlags(config, name, "rpc")

	id := serviceNetConf.ID()

	logger := ctx.Logger.WithOptions(zap.Fields(
		zap.String("service_name", name),
	))

	service := f(id, config, logger, ctx.Discovery)

	if exists {
		for idx := range ctx.Services {
			if ctx.Services[idx].ID == id {
				ctx.Services[idx].service = service
			}
		}
	} else {
		ctx.Services = append(ctx.Services, &serviceRunConfig{
			service: service,
			Name:    name,
			ID:      id,
		})
	}
}

func (ctx *Context) Run(v *viper.Viper) error {
	if allocID := os.Getenv("NOMAD_ALLOC_ID"); allocID == "" {
		fd, err := os.Create(fmt.Sprintf("%s/services_%s.json", path.DataDir(), ctx.ID))
		if err != nil {
			panic(err)
		}
		err = json.NewEncoder(fd).Encode(&ctx.Services)
		if err != nil {
			panic(err)
		}
		fd.Close()
	}
	defer func() {
		ctx.Logger.Sync()
		if r := recover(); r != nil {
			ctx.Logger.Error("panic", zap.String("panic_log", fmt.Sprint(r)))
			os.Exit(9)
		}
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
		logger := ctx.Logger.WithOptions(zap.Fields(
			zap.String("service_name", service.Name),
		))
		identity := ctx.identityCatalog.Get(service.Name, "rpc")
		catalog := discovery.NewServiceCatalog(ctx.identityCatalog, ctx.Discovery)
		err := service.service.Start(service.ID, service.Name, catalog, logger)
		if err != nil {
			logger.Error("failed to start service", zap.Error(err))
		}
		if identity != nil {
			listener := service.service.Serve(identity.BindPort())
			if listener != nil {
				logService(logger, service.Name, identity)
			}
		}
	}
	quit := make(chan struct{})
	go func() {
		defer close(quit)
		<-sigc
		logger.Info("received termination signal")
		go func() {
			<-sigc
			logger.Info("received another termination signal: forcing stop")
			os.Exit(10)
		}()
		for _, service := range ctx.Services {
			logger.Debug(fmt.Sprintf("stopping service %s", service.Name))
			service.service.Shutdown()
			logger.Debug(fmt.Sprintf("stopped service %s", service.Name))
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
	nodeID := makeNodeID(v.GetString("node-id"))
	ctx := &Context{
		ID:              nodeID,
		identityCatalog: identity.NewCatalog(v),
	}
	var logger *zap.Logger
	var err error
	fields := []zap.Field{
		zap.String("node_id", ctx.ID[:8]),
		zap.String("version", Version()),
		zap.Time("started_at", time.Now()),
	}
	if allocID := os.Getenv("NOMAD_ALLOC_ID"); allocID != "" {
		ctx.identityCatalog = identity.NewNomadCatalog()
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
	logger.Debug("starting node")
	ctx.Logger = logger
	if v.GetBool("pprof") {
		go func() {
			fmt.Println("pprof endpoint is running on port 8080")
			http.ListenAndServe(":8080", nil)
		}()
	}
	mesh := createDiscoveryProvider(ctx.ID, v, logger)
	ctx.Discovery = mesh
	if allocID := os.Getenv("NOMAD_ALLOC_ID"); allocID == "" {
		fd, err := os.Open(fmt.Sprintf("%s/services_%s.json", path.DataDir(), ctx.ID))
		if err == nil {
			defer fd.Close()
			err := json.NewDecoder(fd).Decode(&ctx.Services)
			if err != nil {
				panic(err)
			}
		}
	}
	logger.Debug("mesh created")
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

func createDiscoveryProvider(id string, v *viper.Viper, logger *zap.Logger) discovery.DiscoveryAdapter {
	if v.GetString("discovery-provider") == "consul" {
		return discovery.Consul(id, logger)
	}
	if v.GetString("discovery-provider") == "nomad" {
		return discovery.Nomad(id, logger)
	}
	panic("unknown discovery provider")
}
