package broker

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster"
)

// NotifyJoin is called if a peer joins the cluster.
func (b *Broker) NotifyJoin(n *memberlist.Node) {
	var meta cluster.NodeMeta
	err := proto.Unmarshal(n.Meta, &meta)
	if err == nil {
		log.Printf("INFO: node %s (RPC %s) joined the mesh", n.Name, meta.RPCAddr)
	}
}

// NotifyLeave is called if a peer leaves the cluster.
func (b *Broker) NotifyLeave(n *memberlist.Node) {
	log.Printf("INFO: node %s (%s) left the mesh", n.Name, n.Addr.String())
	var meta cluster.NodeMeta
	err := proto.Unmarshal(n.Meta, &meta)
	if err == nil {
		err = b.RPCCaller.Cancel(meta.RPCAddr)
		if err != nil {
			log.Printf("ERROR: failed to close RPC pool toward %s: %v", meta.RPCAddr, err)
		}
	} else {
		log.Printf("ERROR: failed to unmashal peer metadata: %v", err)
	}
	b.onPeerDown(n.Name)
}

// NotifyUpdate is called if a cluster peer gets updated.
func (b *Broker) NotifyUpdate(n *memberlist.Node) {
}
