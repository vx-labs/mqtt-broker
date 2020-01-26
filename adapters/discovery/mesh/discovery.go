package mesh

import (
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/listeners"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/mesh/peers"
	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
)

var (
	ErrStateKeyAlreadySet = errors.New("specified key is already taken")
	ErrNodeNotFound       = errors.New("specified node not found in mesh")
)

type MeshDiscoveryAdapter struct {
	id    string
	layer pb.MembershipAdapter
	peers peers.PeerStore
}

func (m *MeshDiscoveryAdapter) ID() string {
	return m.id
}

type Service struct {
	ID      string
	Address string
}

func NewDiscoveryAdapter(id string, logger *zap.Logger, members pb.MembershipAdapter) *MeshDiscoveryAdapter {
	self := &MeshDiscoveryAdapter{
		id:    id,
		layer: members,
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = os.Getenv("HOSTNAME")
	}
	if hostname == "" {
		hostname = "hostname_not_available"
	}
	peer := pb.Peer{
		ID:       self.id,
		Hostname: hostname,
		Runtime:  runtime.Version(),
		Started:  time.Now().Unix(),
	}
	store, err := peers.NewPeerStore()
	if err != nil {
		panic(err)
	}
	self.peers = store
	err = self.peers.Upsert(&peer)
	if err != nil {
		panic(err)
	}
	self.syncMeta()
	members.OnNodeJoin(func(remoteID string, meta []byte) {
		if id == remoteID {
			return
		}
		p := pb.Peer{}
		err := proto.Unmarshal(meta, &p)
		if err != nil {
			logger.Warn("failed to unmarshal metadata on node join", zap.Error(err))
		}
		err = self.peers.Upsert(&p)
		if err != nil {
			logger.Warn("failed to create peer in local store", zap.Error(err), zap.String("faulty_id", remoteID))
		}
	})
	members.OnNodeLeave(func(remoteID string, _ []byte) {
		if id == remoteID {
			return
		}
		err := self.peers.Delete(remoteID)
		if err != nil {
			logger.Warn("failed to delete peer in local store", zap.Error(err), zap.String("faulty_id", remoteID))
		}
	})
	members.OnNodeUpdate(func(remoteID string, meta []byte) {
		if id == remoteID {
			return
		}
		p := pb.Peer{}
		err := proto.Unmarshal(meta, &p)
		if err != nil {
			logger.Warn("failed to unmarshal metadata on node join", zap.Error(err))
		}
		if self.peers.Exists(remoteID) {
			err = self.peers.Update(remoteID, func(_ pb.Peer) pb.Peer {
				return p
			})
		} else {
			err = self.peers.Upsert(&p)
		}
		if err != nil {
			logger.Warn("failed to update peer in local store", zap.Error(err), zap.String("faulty_id", remoteID))
		}
	})
	go self.oSStatsReporter()

	resolver.Register(NewMeshResolver(self.peers, logger))
	return self
}
func (m *MeshDiscoveryAdapter) Leave() {
	m.layer.Shutdown()
}

func (m *MeshDiscoveryAdapter) RegisterTCPService(id, name, address string) error {
	return m.registerService(id, name, address)
}
func (m *MeshDiscoveryAdapter) RegisterUDPService(id, name, address string) error {
	return m.registerService(id, name, address)
}
func (m *MeshDiscoveryAdapter) RegisterGRPCService(id, name, address string) error {
	return m.registerService(id, name, address)
}
func (m *MeshDiscoveryAdapter) ListenTCP(id, name string, port int, advertizedAddress string) (net.Listener, error) {
	err := m.registerService(id, name, advertizedAddress)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	return &listeners.ServiceTCPListener{
		CloseCallback: func() error {
			return m.UnregisterService(id)
		},
		Listener: listener,
	}, nil
}
func (m *MeshDiscoveryAdapter) ListenUDP(id, name string, port int, advertizedAddress string) (net.PacketConn, error) {
	err := m.registerService(id, name, advertizedAddress)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenPacket("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	return &listeners.ServiceUDPListener{
		CloseCallback: func() error {
			return m.UnregisterService(id)
		},
		Listener: listener,
	}, nil
}
func (m *MeshDiscoveryAdapter) registerService(id, name, address string) error {
	if id == "" {
		return errors.New("empty service id")
	}
	err := m.peers.Update(m.id, func(self pb.Peer) pb.Peer {
		self.HostedServices = append(self.HostedServices, &pb.NodeService{
			ID:             id,
			Name:           name,
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

func (m *MeshDiscoveryAdapter) syncMeta() error {
	self, err := m.peers.ByID(m.id)
	if err == nil {
		payload, err := proto.Marshal(self)
		if err != nil {
			panic(err)
		}
		m.layer.UpdateMetadata(payload)
	}
	return err
}
func (m *MeshDiscoveryAdapter) AddServiceTag(service, key, value string) error {
	err := m.peers.Update(m.id, func(self pb.Peer) pb.Peer {
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
func (m *MeshDiscoveryAdapter) RemoveServiceTag(id string, tag string) error {
	err := m.peers.Update(m.id, func(self pb.Peer) pb.Peer {
		for idx := range self.HostedServices {
			if self.HostedServices[idx].ID == id {
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
func (m *MeshDiscoveryAdapter) UnregisterService(name string) error {
	err := m.peers.Update(m.id, func(self pb.Peer) pb.Peer {
		newServices := []*pb.NodeService{}
		newServiceNames := []string{}
		for _, service := range self.HostedServices {
			if service.Name != name {
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
func (m *MeshDiscoveryAdapter) Members() ([]*pb.Peer, error) {
	set, err := m.peers.All()
	if err != nil {
		return nil, err
	}
	return set.Peers, nil
}
func (m *MeshDiscoveryAdapter) EndpointsByService(name string) ([]*pb.NodeService, error) {
	return m.peers.EndpointsByService(name)
}
