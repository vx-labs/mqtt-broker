package gossip

import (
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/adapters/ap/pb"
	discovery "github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
	"go.uber.org/zap"
)

type Service interface {
	DiscoverEndpoints() ([]*discovery.NodeService, error)
	Address() string
	Name() string
	BindPort() int
	ListenTCP() (net.Listener, error)
	ListenUDP() (net.PacketConn, error)
}

type layer struct {
	id         string
	service    Service
	mlist      *memberlist.Memberlist
	logger     *zap.Logger
	state      pb.APState
	bcastQueue *memberlist.TransmitLimitedQueue
	cancel     chan struct{}
	done       chan struct{}
}

func (m *layer) NotifyMsg(b []byte) {
	if err := m.state.Merge(b, false); err != nil {
		m.logger.Error("failed to merge remote state", zap.Error(err))
		return
	}
}
func (s *layer) Health() string {
	if s.mlist.NumMembers() == 1 {
		return "warning"
	}
	return "ok"
}
func (s *layer) GetBroadcasts(overhead, limit int) [][]byte {
	return s.bcastQueue.GetBroadcasts(overhead, limit)
}
func (m *layer) LocalState(join bool) []byte {
	return m.state.MarshalBinary()
}
func (m *layer) MergeRemoteState(buf []byte, join bool) {
	if err := m.state.Merge(buf, join); err != nil {
		m.logger.Error("failed to merge remote state", zap.Error(err))
	}
}

func (m *layer) join(newHosts []string) error {
	if len(newHosts) == 0 {
		return nil
	}
	count, err := m.mlist.Join(newHosts)
	if err != nil {
		if count == 0 {
			m.logger.Warn("failed to join gossip cluster", zap.Error(err))
			return err
		}
		m.logger.Warn("failed to join some member of cluster", zap.Error(err))
	}
	return nil
}
func (self *layer) discoverPeers() error {
	peers, err := self.service.DiscoverEndpoints()
	if err != nil {
		return err
	}
	addresses := []string{}
	for _, peer := range peers {
		if !self.isNodeKnown(peer.ID) {
			addresses = append(addresses, peer.NetworkAddress)
		}
	}
	if len(addresses) > 0 {
		return self.join(addresses)
	}
	return nil
}
func (self *layer) Shutdown() error {
	self.logger.Info("shutting down gossip state distributer")
	close(self.cancel)
	<-self.done
	err := self.mlist.Leave(5 * time.Second)
	if err != nil {
		return err
	}
	return self.mlist.Shutdown()
}
func (self *layer) isNodeKnown(id string) bool {
	members := self.mlist.Members()
	for _, member := range members {
		if member.Name == id {
			return true
		}
	}
	return false
}
func (self *layer) numMembers() int {
	if self.mlist == nil {
		return 1
	}
	return self.mlist.NumMembers()
}

func (s *layer) NodeMeta(limit int) []byte {
	return nil
}
func NewGossipDistributer(id string, service Service, state pb.APState, logger *zap.Logger) *layer {
	self := &layer{
		id:      id,
		state:   state,
		logger:  logger,
		service: service,
		cancel:  make(chan struct{}),
		done:    make(chan struct{}),
	}

	self.bcastQueue = &memberlist.TransmitLimitedQueue{
		NumNodes:       self.numMembers,
		RetransmitMult: 3,
	}
	go func() {
		defer func() { close(self.done) }()
		for {
			select {
			case event := <-state.Events():
				self.bcastQueue.QueueBroadcast(simpleBroadcast(event))
			case <-self.cancel:
				return
			}
		}
	}()

	config := memberlist.DefaultLANConfig()
	address, portStr, err := net.SplitHostPort(service.Address())
	if err != nil {
		panic(err)
	}
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		panic(err)
	}
	config.AdvertiseAddr = address
	config.AdvertisePort = int(port)
	config.BindPort = service.BindPort()
	config.Name = id
	config.Delegate = self
	if os.Getenv("ENABLE_MEMBERLIST_LOG") != "true" {
		config.LogOutput = ioutil.Discard
	}
	config.Transport, err = newNetTransport(service, logger)
	if err != nil {
		panic(err)
	}
	list, err := memberlist.Create(config)
	if err != nil {
		panic(err)
	}
	self.mlist = list
	go func() {
		retryTicker := time.NewTicker(5 * time.Second)
		defer retryTicker.Stop()
		for range retryTicker.C {
			err := self.discoverPeers()
			if err == nil {
				return
			}
			logger.Warn("failed to join gossip cluster members", zap.Error(err))
		}
	}()
	return self
}
