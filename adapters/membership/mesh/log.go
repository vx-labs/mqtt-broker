package mesh

import (
	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

// NotifyJoin is called if a peer joins the cluster.
func (b *GossipMembershipAdapter) NotifyJoin(n *memberlist.Node) {
	b.logger.Debug("node joined", zap.String("new_node_id", n.Name))
	if b.onNodeJoin != nil && n.Meta != nil {
		b.onNodeJoin(n.Name, b.decodeMeta(n.Meta))
	}
}

// NotifyLeave is called if a peer leaves the cluster.
func (b *GossipMembershipAdapter) NotifyLeave(n *memberlist.Node) {
	b.logger.Debug("node left", zap.String("left_node_id", n.Name))
	if b.onNodeLeave != nil && n.Meta != nil {
		b.onNodeLeave(n.Name, b.decodeMeta(n.Meta))
	}
}

// NotifyUpdate is called if a cluster peer gets updated.
func (b *GossipMembershipAdapter) NotifyUpdate(n *memberlist.Node) {
	b.logger.Debug("node updated", zap.String("updated_node_id", n.Name))
	if b.onNodeUpdate != nil && n.Meta != nil {
		b.onNodeUpdate(n.Name, b.decodeMeta(n.Meta))
	}
}
