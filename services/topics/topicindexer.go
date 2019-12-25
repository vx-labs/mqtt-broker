package topics

import (
	"github.com/vx-labs/mqtt-broker/services/topics/pb"
	"github.com/vx-labs/mqtt-broker/services/topics/tree"
)

type topicIndexer struct {
	root *tree.Node
}

func TenantTopicIndexer() *topicIndexer {
	return &topicIndexer{
		root: tree.NewNode("_root", "_all"),
	}
}

func (t *topicIndexer) Lookup(tenant string, pattern []byte) (pb.RetainedMessageSet, error) {
	var vals pb.RetainedMessageSet
	topic := pb.NewTopic(pattern)
	t.root.Apply(tenant, topic, func(node *tree.Node) bool {
		if len(node.Message.Payload) > 0 {
			vals = append(vals, node.Message)
		}
		return false
	})
	return vals, nil
}

func (s *topicIndexer) Index(message pb.RetainedMessage) error {
	topic := pb.NewTopic(message.GetTopic())
	node := s.root.Upsert(message.GetTenant(), topic)
	node.Message = message
	return nil
}
