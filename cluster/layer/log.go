package layer

import (
	"github.com/hashicorp/memberlist"
)

// NotifyJoin is called if a peer joins the cluster.
func (b *layer) NotifyJoin(n *memberlist.Node) {
	if b.onNodeJoin != nil {
		b.onNodeJoin(n.Name, b.decodeMeta(n.Meta))
	}
}

// NotifyLeave is called if a peer leaves the cluster.
func (b *layer) NotifyLeave(n *memberlist.Node) {
	if b.onNodeLeave != nil {
		b.onNodeLeave(n.Name, b.decodeMeta(n.Meta))
	}
}

// NotifyUpdate is called if a cluster peer gets updated.
func (b *layer) NotifyUpdate(n *memberlist.Node) {
	if b.onNodeUpdate != nil {
		b.onNodeUpdate(n.Name, b.decodeMeta(n.Meta))
	}
}
