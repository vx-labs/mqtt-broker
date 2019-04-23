package topics

type topicIndexer struct {
	root *Node
}

func TenantTopicIndexer() *topicIndexer {
	return &topicIndexer{
		root: NewNode("_root", "_all"),
	}
}

func (t *topicIndexer) Lookup(tenant string, pattern []byte) (RetainedMessageSet, error) {
	var vals RetainedMessageSet
	topic := NewTopic(pattern)
	t.root.Apply(tenant, topic, func(node *Node) bool {
		if len(node.Message.Payload) > 0 {
			vals = append(vals, node.Message)
		}
		return false
	})
	return vals, nil
}

func (s *topicIndexer) Index(message RetainedMessage) error {
	topic := NewTopic(message.GetTopic())
	node := s.root.Upsert(message.GetTenant(), topic)
	node.Message = message
	return nil
}
