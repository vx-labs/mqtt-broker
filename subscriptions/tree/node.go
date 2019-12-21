package tree

import (
	"sync/atomic"
	"unsafe"

	"github.com/vx-labs/mqtt-broker/subscriptions/pb"
)

type Node struct {
	pattern []byte
	tenant  string
	data    []*pb.Subscription
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

func (n *Node) AddSubscription(tenant string, subscription *pb.Subscription) *Node {
	newNode := NewNode(tenant, n.pattern)
	newNode.data = append(n.data, subscription)
	newNode.inode = n.inode
	return newNode
}
func (n *Node) DelSubscription(id string) *Node {
	newNode := NewNode(n.tenant, n.pattern)
	for _, subscription := range n.data {
		if subscription.ID != id {
			newNode.data = append(newNode.data, subscription)
		}
	}
	newNode.inode = n.inode
	return newNode
}
