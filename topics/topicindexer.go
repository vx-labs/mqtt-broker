package topics

type topicIndexer struct {
	root *Node
}

func TenantTopicIndexer() *topicIndexer {
	return &topicIndexer{
		root: NewNode("_root", "_all"),
	}
}

func (t *topicIndexer) Lookup(tenant string, pattern []byte) (RetainedMessageList, error) {
	var vals RetainedMessageList
	topic := NewTopic(pattern)
	t.root.Apply(tenant, topic, func(node *Node) bool {
		if node.Message != nil {
			vals.RetainedMessages = append(vals.RetainedMessages, node.Message)
		}
		return false
	})
	return vals, nil
}

func (s *topicIndexer) Index(message *RetainedMessage) error {
	topic := NewTopic(message.GetTopic())
	node := s.root.Upsert(message.GetTenant(), topic)
	node.Message = message
	return nil
}
