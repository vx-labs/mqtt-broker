package broker

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/broker/cluster"
)

// NotifyJoin is called if a peer joins the cluster.
func (b *Broker) NotifyJoin(n *memberlist.Node) {
	log.Printf("INFO: node %s (%s) joined the mesh", n.Name, n.Addr.String())

}

// NotifyLeave is called if a peer leaves the cluster.
func (b *Broker) NotifyLeave(n *memberlist.Node) {
	log.Printf("INFO: node %s (%s) left the mesh", n.Name, n.Addr.String())
	var meta cluster.NodeMeta
	err := proto.Unmarshal(n.Meta, &meta)
	if err == nil {
		b.RPCCaller.Cancel(meta.RPCAddr)
	}
	b.onPeerDown(n.Name)
}

// NotifyUpdate is called if a cluster peer gets updated.
func (b *Broker) NotifyUpdate(n *memberlist.Node) {
}
