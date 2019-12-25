package tree

import (
	"fmt"
	"sync"

	"github.com/vx-labs/mqtt-broker/services/topics/pb"
)

var (
	ErrNodeNotFound = fmt.Errorf("node not found")
)

type Node struct {
	Id       string
	Message  pb.RetainedMessage
	Tenant   string
	mutex    sync.RWMutex
	children map[string]*Node
}

func NewNode(id, tenant string) *Node {
	n := &Node{
		Id:       id,
		Tenant:   tenant,
		children: map[string]*Node{},
	}
	return n
}
func (n *Node) Child(id string) (*Node, bool) {
	n.mutex.RLock()
	c, ok := n.children[id]
	n.mutex.RUnlock()
	return c, ok
}
func (n *Node) Children(tenant string) map[string]*Node {
	n.mutex.RLock()
	children := make(map[string]*Node, len(n.children))
	for idx, c := range n.children {
		if c.Tenant == tenant {
			children[idx] = c
		}
	}
	n.mutex.RUnlock()
	return children
}
func (n *Node) AddChild(id, tenant string) *Node {
	n.mutex.Lock()
	c := NewNode(id, tenant)
	n.children[id] = c
	n.mutex.Unlock()
	return c
}

func (n *Node) Apply(tenant string, topic *pb.Topic, f func(*Node) bool) {
	var (
		token string
		read  bool
	)
	topic, token, read = topic.Chop()
	switch token {
	case "#":
		for _, child := range n.Children(tenant) {
			if f(child) {
				return
			}
			child.Apply(tenant, pb.NewTopic([]byte("#")), f)
		}
	case "+":
		for _, child := range n.Children(tenant) {
			if read {
				child.Apply(tenant, topic, f)
			} else {
				if f(child) {
					return
				}
			}
		}
	default:
		child, ok := n.Child(token)
		if !ok {
			return
		}
		if read {
			child.Apply(tenant, topic, f)
		} else {
			if f(child) {
				return
			}
		}
	}
}
func (n *Node) Upsert(tenant string, topic *pb.Topic) *Node {
	var (
		token string
		read  bool
	)
	topic, token, read = topic.Chop()
	child, ok := n.Child(token)
	if !ok {
		child = n.AddChild(token, tenant)
	}
	if read {
		return child.Upsert(tenant, topic)
	}
	return child
}
