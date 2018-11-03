package subscriptions

import (
	"sync/atomic"
	"unsafe"
)

type Node struct {
	pattern []byte
	tenant  string
	data    SubscriptionList
	inode   *INode
}

func NewNode(tenant string, pattern []byte) *Node {
	return &Node{
		tenant:  tenant,
		pattern: pattern,
		inode:   NewINode(),
	}
}

func (n *Node) casINode(old, cur unsafe.Pointer) bool {
	dest := (*unsafe.Pointer)(unsafe.Pointer(&(n.inode)))
	return atomic.CompareAndSwapPointer(dest, old, cur)
}

func (n *Node) AddSubscription(tenant string, subscription *Subscription) *Node {
	newNode := NewNode(tenant, n.pattern)
	newNode.data.Subscriptions = append(n.data.Subscriptions, subscription)
	newNode.inode = n.inode
	return newNode
}
func (n *Node) DelSubscription(id string) *Node {
	newNode := NewNode(n.tenant, n.pattern)
	for _, subscription := range n.data.Subscriptions {
		if subscription.ID != id {
			newNode.data.Subscriptions = append(newNode.data.Subscriptions, subscription)
		}
	}
	newNode.inode = n.inode
	return newNode
}
