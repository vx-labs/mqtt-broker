package peer

import (
	fmt "fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/vx-labs/mqtt-broker/identity"
	"github.com/weaveworks/mesh"
)

func nameFromID(id string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s", id[0:2], id[2:4], id[4:6], id[6:8], id[8:10], id[10:12])
}

type Peer struct {
	name      mesh.PeerName
	id        identity.Identity
	gossip    mesh.Gossip
	router    *mesh.Router
	onUnicast func([]byte)
}

func (p *Peer) Gossip() (complete mesh.GossipData) {
	return nil
}
func (p *Peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	return nil, nil
}
func (p *Peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	return nil, nil
}
func (p *Peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	p.onUnicast(buf)
	return nil
}
func (p *Peer) Name() mesh.PeerName {
	return p.name
}

func (p *Peer) Router() *mesh.Router {
	return p.router
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
	p.router.ConnectionMaker.InitiateConnections(hosts, false)
}
func (p *Peer) Send(recipient mesh.PeerName, payload []byte) {
	p.gossip.GossipUnicast(recipient, payload)
}
func (p *Peer) Start() {
	p.router.Start()
}
func NewPeer(id identity.Identity, onLost func(mesh.PeerName), onUnicast func([]byte)) *Peer {
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
		name:      name,
		id:        id,
		onUnicast: onUnicast,
	}
	gossip, err := router.NewGossip("mqtt-broker", self)
	if err != nil {
		log.Fatal(err)
	}
	router.Peers.OnGC(func(peer *mesh.Peer) {
		onLost(peer.Name)
	})
	self.gossip = gossip
	self.router = router
	return self
}
