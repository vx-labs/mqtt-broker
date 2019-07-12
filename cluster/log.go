package cluster

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
)

// NotifyJoin is called if a peer joins the cluster.
func (b *layer) NotifyJoin(n *memberlist.Node) {
	var meta pb.NodeMeta
	err := proto.Unmarshal(n.Meta, &meta)
	if err == nil {
		if b.id != n.Name {
			log.Printf("INFO: service/%s: node %s joined the mesh", b.name, n.Name)
		}
		if b.onNodeJoin != nil {
			b.onNodeJoin(n.Name, meta)
		}
	}

}

// NotifyLeave is called if a peer leaves the cluster.
func (b *layer) NotifyLeave(n *memberlist.Node) {
	var meta pb.NodeMeta
	err := proto.Unmarshal(n.Meta, &meta)
	if err == nil {
		if b.id != n.Name {
			log.Printf("INFO: service/%s: node %s left the mesh", b.name, n.Name)
		}
		if b.onNodeLeave != nil {
			b.onNodeLeave(n.Name, meta)
		}
	}
}

// NotifyUpdate is called if a cluster peer gets updated.
func (b *layer) NotifyUpdate(n *memberlist.Node) {
}
