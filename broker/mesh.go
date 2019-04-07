package broker

import (
	"log"

	"github.com/hashicorp/memberlist"
)

// NotifyJoin is called if a peer joins the cluster.
func (b *Broker) NotifyJoin(n *memberlist.Node) {
	log.Printf("INFO: node %s (%s) joined the mesh", n.Name, n.Addr.String())

}

// NotifyLeave is called if a peer leaves the cluster.
func (b *Broker) NotifyLeave(n *memberlist.Node) {
	log.Printf("INFO: node %s (%s) left the mesh", n.Name, n.Addr.String())
	if b.Peers != nil {
		b.Peers.Delete(n.Name)
	}
}

// NotifyUpdate is called if a cluster peer gets updated.
func (b *Broker) NotifyUpdate(n *memberlist.Node) {
}
