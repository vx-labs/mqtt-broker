package discovery

import (
	"errors"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"

	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/layer"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/pool"
	"github.com/vx-labs/mqtt-broker/cluster/types"
)

var (
	ErrStateKeyAlreadySet = errors.New("specified key is already taken")
	ErrNodeNotFound       = errors.New("specified node not found in mesh")
)

type discoveryLayer struct {
	id        string
	rpcCaller *pool.Caller
	layer     types.GossipServiceLayer
	peers     peers.PeerStore
}

func (m *discoveryLayer) ID() string {
	return m.id
}
func (m *discoveryLayer) Peers() peers.PeerStore {
	return m.peers
}

type Service struct {
	ID      string
	Address string
}

func NewDiscoveryLayer(logger *zap.Logger, userConfig config.Config) *discoveryLayer {
	self := &discoveryLayer{
		id:        userConfig.ID,
		rpcCaller: pool.NewCaller(),
	}
	userConfig.OnNodeLeave = func(id string, meta pb.NodeMeta) {
		for _, service := range meta.Services {
			self.rpcCaller.Cancel(service.NetworkAddress)
		}
		self.peers.Delete(id)
	}
	self.layer = layer.NewGossipLayer("cluster", logger, userConfig, pb.NodeMeta{
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
	go self.oSStatsReporter()
	return self
}
func (m *discoveryLayer) Leave() {
	m.layer.Leave()
}
func (m *discoveryLayer) Join(peers []string) error {
	return m.layer.Join(peers)
}

func (m *discoveryLayer) ServiceName() string {
	return "cluster"
}
func (m *discoveryLayer) RegisterService(name, address string) error {
	return m.peers.Update(m.id, func(self peers.Peer) peers.Peer {
		self.HostedServices = append(self.HostedServices, &pb.NodeService{
			ID:             name,
			NetworkAddress: address,
			Peer:           m.id,
		})
		self.Services = append(self.Services, name)
		return self
	})
}
func (m *discoveryLayer) Health() string {
	if len(m.layer.Members()) > 1 {
		return "ok"
	}
	return "warning"
}
