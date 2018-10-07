package peer

import (
	fmt "fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/vx-labs/mqtt-broker/broker/peer/state"
	"github.com/vx-labs/mqtt-broker/identity"
	"github.com/weaveworks/mesh"
)

type State interface {
	mesh.GossipData
	GossipData() mesh.GossipData
	MergeDelta(buf []byte) (delta mesh.GossipData)
	Add(ev string)
	Remove(ev string)
	Iterate(f func(ev string, added, deleted bool) error) error
}

func nameFromID(id string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s", id[0:2], id[2:4], id[4:6], id[6:8], id[8:10], id[10:12])
}

type Peer struct {
	name           mesh.PeerName
	id             identity.Identity
	state          State
	gossip         mesh.Gossip
	router         *mesh.Router
	onUnicast      func([]byte)
	broadcastTimer chan struct{}
}

func (p *Peer) Gossip() (complete mesh.GossipData) {
	st := p.state.GossipData()
	return st
}
func (p *Peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	st := p.state.MergeDelta(buf)
	return st, nil
}
func (p *Peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	return p.state.MergeDelta(buf), nil
}
func (p *Peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	p.onUnicast(buf)
	return nil
}
func (p *Peer) Name() mesh.PeerName {
	return p.name
}
func (p *Peer) Add(ev string) {
	p.state.Add(ev)
	p.triggerBroadcast()
}
func (p *Peer) Del(ev string) {
	p.state.Remove(ev)
	p.triggerBroadcast()
}
func (p *Peer) triggerBroadcast() {
	select {
	case token := <-p.broadcastTimer:
		p.gossip.GossipBroadcast(p.state)
		go func() {
			time.Sleep(100 * time.Millisecond)
			p.broadcastTimer <- token
		}()
	default:
	}
}
func (p *Peer) Members() []mesh.PeerName {
	peers := p.router.Peers.Descriptions()
	members := make([]mesh.PeerName, len(peers))
	for _, peer := range peers {
		members = append(members, peer.Name)
	}
	return members
}
func (p *Peer) Join(hosts []string) {
	p.router.ConnectionMaker.InitiateConnections(hosts, true)
}
func (p *Peer) Send(recipient mesh.PeerName, payload []byte) {
	p.gossip.GossipUnicast(recipient, payload)
}
func NewPeer(id identity.Identity, onAdd, onDel func(ev string), onLost func(mesh.PeerName), onUnicast func([]byte)) *Peer {
	name, err := mesh.PeerNameFromString(nameFromID(id.ID()))
	if err != nil {
		log.Fatal(err)
	}
	router, err := mesh.NewRouter(mesh.Config{
		Host:               id.Private().Host(),
		Port:               id.Private().Port(),
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		ConnLimit:          64,
		PeerDiscovery:      true,
		TrustedSubnets:     []*net.IPNet{},
	}, name, id.Public().String(), mesh.NullOverlay{}, log.New(ioutil.Discard, "mesh/router: ", 0))
	if err != nil {
		log.Fatal(err)
	}
	self := &Peer{
		name:           name,
		id:             id,
		state:          state.New(name, onAdd, onDel),
		onUnicast:      onUnicast,
		broadcastTimer: make(chan struct{}, 1),
	}
	self.broadcastTimer <- struct{}{}
	gossip, err := router.NewGossip("mqtt-broker", self)
	if err != nil {
		log.Fatal(err)
	}
	router.Peers.OnGC(func(peer *mesh.Peer) {
		onLost(peer.Name)
	})
	router.Start()
	self.gossip = gossip
	self.router = router
	return self
}
