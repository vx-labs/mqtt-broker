package discovery

import (
	"errors"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/cluster/config"
	"github.com/vx-labs/mqtt-broker/cluster/layer"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
)

var (
	ErrStateKeyAlreadySet = errors.New("specified key is already taken")
	ErrNodeNotFound       = errors.New("specified node not found in mesh")
)

type discoveryLayer struct {
	id    string
	layer types.GossipServiceLayer
	peers peers.PeerStore
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
		id: userConfig.ID,
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = os.Getenv("HOSTNAME")
	}
	if hostname == "" {
		hostname = "hostname_not_available"
	}
	peer := peers.Peer{
		Metadata: pb.Metadata{
			ID:       self.id,
			Hostname: hostname,
			Runtime:  runtime.Version(),
			Started:  time.Now().Unix(),
		},
	}
	payload, err := proto.Marshal(&peer)
	if err != nil {
		panic(err)
	}
	store, err := peers.NewPeerStore()
	if err != nil {
		panic(err)
	}
	self.peers = store
	self.peers.Upsert(peer)

	userConfig.OnNodeJoin = func(id string, meta []byte) {
		if id == userConfig.ID {
			return
		}
		p := peers.Peer{}
		err := proto.Unmarshal(meta, &p)
		if err != nil {
			logger.Warn("failed to unmarshal metadata on node join", zap.Error(err))
		}
		err = self.peers.Upsert(p)
		if err != nil {
			logger.Warn("failed to update peer in local store", zap.Error(err))
		}
	}
	userConfig.OnNodeLeave = func(id string, _ []byte) {
		if id == userConfig.ID {
			return
		}
		err := self.peers.Delete(id)
		if err != nil {
			logger.Warn("failed to delete peer in local store", zap.Error(err))
		}
	}
	userConfig.OnNodeUpdate = func(id string, meta []byte) {
		if id == userConfig.ID {
			return
		}
		p := peers.Peer{}
		err := proto.Unmarshal(meta, &p)
		if err != nil {
			logger.Warn("failed to unmarshal metadata on node join", zap.Error(err))
		}
		err = self.peers.Update(id, func(_ peers.Peer) peers.Peer {
			return p
		})
		if err != nil {
			logger.Warn("failed to update peer in local store", zap.Error(err))
		}
	}
	self.layer = layer.NewGossipLayer("cluster", logger, userConfig, payload)
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
	err := m.peers.Update(m.id, func(self peers.Peer) peers.Peer {
		self.HostedServices = append(self.HostedServices, &pb.NodeService{
			ID:             name,
			NetworkAddress: address,
			Peer:           m.id,
		})
		self.Services = append(self.Services, name)
		return self
	})
	if err == nil {
		m.syncMeta()
	}
	return err
}

func (m *discoveryLayer) syncMeta() error {
	self, err := m.peers.ByID(m.id)
	if err == nil {
		payload, err := proto.Marshal(&self)
		if err != nil {
			panic(err)
		}
		m.layer.UpdateMeta(payload)
	}
	return err
}
func (m *discoveryLayer) AddServiceTag(service, key, value string) error {
	err := m.peers.Update(m.id, func(self peers.Peer) peers.Peer {
		for idx := range self.HostedServices {
			if self.HostedServices[idx].ID == service {
				found := false
				for _, tag := range self.HostedServices[idx].Tags {
					found = true
					if tag.Key == key {
						tag.Value = value
					}
					break
				}
				if !found {
					self.HostedServices[idx].Tags = append(self.HostedServices[idx].Tags, &pb.ServiceTag{
						Key:   key,
						Value: value,
					})
				}
				break
			}
		}
		return self
	})
	if err == nil {
		return m.syncMeta()
	}
	return err
}
func (m *discoveryLayer) RemoveServiceTag(name string, tag string) error {
	err := m.peers.Update(m.id, func(self peers.Peer) peers.Peer {
		for idx := range self.HostedServices {
			if self.HostedServices[idx].ID == name {
				dirty := false
				tags := []*pb.ServiceTag{}
				for _, currentTag := range self.HostedServices[idx].Tags {
					if currentTag.Key != tag {
						dirty = true
						tags = append(tags, currentTag)
					}
				}
				if dirty {
					self.HostedServices[idx].Tags = tags
				}
				break
			}
		}
		return self
	})
	if err == nil {
		return m.syncMeta()
	}
	return err
}
func (m *discoveryLayer) UnregisterService(name string) error {
	err := m.peers.Update(m.id, func(self peers.Peer) peers.Peer {
		newServices := []*pb.NodeService{}
		newServiceNames := []string{}
		for _, service := range self.HostedServices {
			if service.ID != name {
				newServices = append(newServices, service)
			}
		}
		for _, service := range self.Services {
			if service != name {
				newServiceNames = append(newServiceNames, service)
			}
		}
		self.HostedServices = newServices
		self.Services = newServiceNames
		return self
	})
	if err == nil {
		return m.syncMeta()
	}
	return err
}
func (m *discoveryLayer) Health() string {
	if len(m.layer.Members()) > 1 {
		return "ok"
	}
	return "warning"
}
