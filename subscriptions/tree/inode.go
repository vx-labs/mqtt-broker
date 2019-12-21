package tree

import (
	"bytes"
	"errors"

	"github.com/vx-labs/mqtt-broker/crdt"

	"github.com/vx-labs/mqtt-broker/subscriptions/pb"
	"github.com/vx-labs/mqtt-broker/subscriptions/topic"
)

type INode struct {
	nodes []*Node
}

func NewINode() *INode {
	return &INode{}
}

func (d *INode) AddNode(node *Node) {
	d.nodes = append(d.nodes, node)
}

func findChild(d *INode, tenant string, pattern []byte) *Node {
	for _, node := range d.nodes {
		if tenant == node.tenant && bytes.Compare(node.pattern, pattern) == 0 {
			return node
		}
	}
	node := NewNode(tenant, pattern)
	d.AddNode(node)
	return node
}

func (d *INode) Insert(topic topic.Topic, tenant string, subscription *pb.Subscription) {
	inode := d
	var (
		token []byte
		ok    bool
	)
	for {
		topic, token, ok = topic.Chop()
		node := findChild(inode, tenant, token)
		if ok {
			inode = node.inode
		} else {
			node = node.AddSubscription(tenant, subscription)
			for idx := range inode.nodes {
				if inode.nodes[idx].tenant == tenant &&
					bytes.Compare(node.pattern, inode.nodes[idx].pattern) == 0 {
					inode.nodes[idx] = node
					return
				}
			}
			break
		}
	}
}
func (d *INode) Remove(tenant, id string, topic topic.Topic) error {
	topic, token, ok := topic.Chop()
	for idx, node := range d.nodes {
		if bytes.Equal(token, node.pattern) {
			if !ok {
				target := []*pb.Subscription{}
				for _, sub := range node.data {
					if sub.ID == id {
						target = append(target, sub)
					}
				}
				if len(target) == 0 {
					return errors.New("subscription not found in index")
				}
				d.nodes[idx] = node.DelSubscription(id)
				return nil
			} else {
				return node.inode.Remove(tenant, id, topic)
			}
		}
	}
	return errors.New("subscription not found in index")
}

func (d *INode) Select(tenant string, set []*pb.Subscription, topic topic.Topic) []*pb.Subscription {
	if set == nil {
		set = []*pb.Subscription{}
	}
	topic, token, ok := topic.Chop()
	for _, node := range d.nodes {
		if node.tenant == tenant && matchPattern(token, node.pattern) {
			if !ok || (len(node.pattern) == 1 && node.pattern[0] == '#') {
				for _, sub := range node.data {
					if crdt.IsEntryAdded(sub) {
						set = append(set, node.data...)
					}
				}
			} else {
				set = node.inode.Select(tenant, set, topic)
			}
		}
	}
	return set
}

func matchPattern(token, pattern []byte) bool {
	if len(pattern) == 1 {
		switch pattern[0] {
		case '+':
			fallthrough
		case '#':
			return true
		}
	}
	return bytes.Compare(token, pattern) == 0
}
