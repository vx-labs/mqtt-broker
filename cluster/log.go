package cluster

import (
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"go.uber.org/zap"
)

// NotifyJoin is called if a peer joins the cluster.
func (b *layer) NotifyJoin(n *memberlist.Node) {
	var meta pb.NodeMeta
	err := proto.Unmarshal(n.Meta, &meta)
	if err == nil {
		if b.id != n.Name {
			b.logger.Info("node joined the mesh", zap.String("peer_id", n.Name))
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
			b.logger.Info("node left the mesh", zap.String("peer_id", n.Name))
		}
		if b.onNodeLeave != nil {
			b.onNodeLeave(n.Name, meta)
		}
	}
}

// NotifyUpdate is called if a cluster peer gets updated.
func (b *layer) NotifyUpdate(n *memberlist.Node) {
}
