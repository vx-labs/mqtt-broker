package cluster

import (
	"errors"
	fmt "fmt"
	"log"
	"os"
	"runtime"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/pool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

var (
	ErrStateKeyAlreadySet = errors.New("specified key is already taken")
	ErrNodeNotFound       = errors.New("specified node not found in mesh")
)

type memberlistMesh struct {
	id        string
	rpcCaller *pool.Caller
	layer     Layer
	peers     peers.PeerStore
}

type cachedState struct {
	data []byte
}

func (c *cachedState) MarshalBinary() []byte {
	return c.data
}

func (c *cachedState) Merge(b []byte, full bool) error {
	if full {
		c.data = b
	}
	return nil
}

func (m *memberlistMesh) DialService(name string) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("mesh:///%s", name),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithInsecure(), grpc.WithAuthority(name), grpc.WithBalancerName("failover"))
}
func (m *memberlistMesh) DialAddress(service, id string, f func(*grpc.ClientConn) error) error {
	return m.rpcCaller.Call(fmt.Sprintf("%s+%s", service, id), f)
}

func (m *memberlistMesh) ID() string {
	return m.id
}
func (m *memberlistMesh) Peers() peers.PeerStore {
	return m.peers
}

type Service struct {
	ID      string
	Address string
}

func New(userConfig Config) *memberlistMesh {
	self := &memberlistMesh{
		id:        userConfig.ID,
		rpcCaller: pool.NewCaller(),
	}
	userConfig.onNodeLeave = func(id string, meta pb.NodeMeta) {
		for _, service := range meta.Services {
			self.rpcCaller.Cancel(service.NetworkAddress)
		}
		self.peers.Delete(id)
		log.Printf("INFO: deleted peer %s from discovery store", id)
	}
	self.layer = NewLayer("cluster", userConfig, pb.NodeMeta{
		ID: userConfig.ID,
	})
	store, err := peers.NewPeerStore(self.layer)
	self.layer.AddState("", store)
	if err != nil {
		panic(err)
	}
	self.peers = store
	hostname, err := os.Hostname()
	if err != nil {
		hostname = os.Getenv("HOSTNAME")
	}
	if hostname == "" {
		hostname = "hostname_not_available"
	}

	self.peers.Upsert(peers.Peer{
		Metadata: pb.Metadata{
			ID:       self.id,
			Hostname: hostname,
			Runtime:  runtime.Version(),
			Started:  time.Now().Unix(),
		},
	})
	resolver.Register(newResolver(self.peers))
	resolver.Register(newIDResolver(self.peers))
	go self.oSStatsReporter()
	return self
}

func (m *memberlistMesh) Leave() {
	m.layer.Leave()
}
func (m *memberlistMesh) Join(peers []string) {
	m.layer.Join(peers)
}

func (m *memberlistMesh) RegisterService(name, address string) error {
	self, err := m.peers.ByID(m.id)
	if err != nil {
		return err
	}
	self.HostedServices = append(self.HostedServices, &pb.NodeService{
		ID:             name,
		NetworkAddress: address,
		Peer:           m.id,
	})
	self.Services = append(self.Services, name)
	log.Printf("INFO: registering service %s on %s", name, address)
	return m.peers.Upsert(self)
}
func (m *memberlistMesh) Health() string {
	if len(m.layer.Members()) > 1 {
		return "ok"
	}
	return "warning"
}
