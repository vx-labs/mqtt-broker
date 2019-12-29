package mesh

import (
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"
	"os"
	"time"

	discovery "github.com/vx-labs/mqtt-broker/adapters/discovery/pb"

	"github.com/hashicorp/memberlist"

	"go.uber.org/zap"
)

type DiscoveryService interface {
	DiscoverEndpoints() ([]*discovery.NodeService, error)
	Address() string
	AdvertisedHost() string
	AdvertisedPort() int
	BindPort() int
}

type Channel interface {
	Broadcast(b []byte)
	BroadcastFullState(b []byte)
}

type GossipMembershipAdapter struct {
	id           string
	mlist        *memberlist.Memberlist
	logger       *zap.Logger
	meta         []byte
	onNodeJoin   func(id string, meta []byte)
	onNodeLeave  func(id string, meta []byte)
	onNodeUpdate func(id string, meta []byte)
}

func (m *GossipMembershipAdapter) Members() []*memberlist.Node {
	return m.mlist.Members()
}

func (m *GossipMembershipAdapter) OnNodeJoin(f func(id string, meta []byte)) {
	m.onNodeJoin = f
	members := m.mlist.Members()
	for idx := range members {
		n := members[idx]
		if n.Meta != nil {
			f(n.Name, m.decodeMeta(n.Meta))
		}
	}
}
func (m *GossipMembershipAdapter) OnNodeLeave(f func(id string, meta []byte)) {
	m.onNodeLeave = f
}
func (m *GossipMembershipAdapter) OnNodeUpdate(f func(id string, meta []byte)) {
	m.onNodeUpdate = f
}

func (s *GossipMembershipAdapter) Health() string {
	if s.mlist.NumMembers() == 1 {
		return "warning"
	}
	return "ok"
}
func (s *GossipMembershipAdapter) decodeMeta(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	r := bytes.NewBuffer(b)
	uncompressed, err := zlib.NewReader(r)
	if err != nil {
		panic(err)
	}
	out := bytes.NewBuffer(nil)
	_, err = io.Copy(out, uncompressed)
	if err != nil {
		panic(err)
	}
	return out.Bytes()
}
func (s *GossipMembershipAdapter) NodeMeta(limit int) []byte {
	if len(s.meta) == 0 {
		return nil
	}
	b := bytes.NewBuffer(nil)
	w := zlib.NewWriter(b)
	_, err := w.Write(s.meta)
	if err != nil {
		panic(err)
	}
	err = w.Close()
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func (m *GossipMembershipAdapter) Join(hosts []string) error {
	if len(hosts) == 0 {
		return nil
	}
	count, err := m.mlist.Join(hosts)
	if err != nil {
		if count == 0 {
			if m.mlist.NumMembers() == 1 {
				m.logger.Debug("failed to join membership cluster", zap.Error(err))
				return err
			}
		}
		m.logger.Warn("failed to join some hosts", zap.Error(err))
	}
	return nil
}
func (self *GossipMembershipAdapter) UpdateMetadata(meta []byte) {
	self.meta = meta
	self.mlist.UpdateNode(5 * time.Second)
}
func (self *GossipMembershipAdapter) Shutdown() error {
	err := self.mlist.Leave(5 * time.Second)
	if err != nil {
		return err
	}
	return self.mlist.Shutdown()
}
func (self *GossipMembershipAdapter) isNodeKnown(id string) bool {
	members := self.mlist.Members()
	for _, member := range members {
		if member.Name == id {
			return true
		}
	}
	return false
}
func (self *GossipMembershipAdapter) numMembers() int {
	if self.mlist == nil {
		return 1
	}
	return self.mlist.NumMembers()
}
func (self *GossipMembershipAdapter) MemberCount() int {
	return self.numMembers()
}

func NewMesh(id string, logger *zap.Logger, service DiscoveryService) *GossipMembershipAdapter {
	self := &GossipMembershipAdapter{
		id:     id,
		meta:   []byte{},
		logger: logger,
	}

	config := memberlist.DefaultLANConfig()
	config.AdvertiseAddr = service.AdvertisedHost()
	config.AdvertisePort = service.AdvertisedPort()
	config.BindPort = service.BindPort()
	config.Name = id
	config.Delegate = self
	config.Events = self
	if os.Getenv("ENABLE_MEMBERLIST_LOG") != "true" {
		config.LogOutput = ioutil.Discard
	}
	list, err := memberlist.Create(config)
	if err != nil {
		panic(err)
	}
	self.mlist = list
	startJoiner(service, self, logger)

	return self
}
