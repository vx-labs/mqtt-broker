package tree

import (
	"sync/atomic"
	"unsafe"

	"github.com/vx-labs/mqtt-broker/services/subscriptions/pb"
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
	size := 1
	for _, s := range n.data {
		if s.ID != subscription.ID {
			size++
		}
	}
	newNode.data = make([]*pb.Subscription, size)
	idx := 0
	for curIdx := range n.data {
		s := n.data[curIdx]
		if s.ID != subscription.ID {
			newNode.data[idx] = s
			idx++
		}
	}
	newNode.data[idx] = subscription
	newNode.inode = n.inode
	return newNode
}
func (n *Node) DelSubscription(id string) *Node {
	newNode := NewNode(n.tenant, n.pattern)
	size := 0
	for _, s := range n.data {
		if s.ID != id {
			size++
		}
	}
	newNode.data = make([]*pb.Subscription, size)
	idx := 0
	for curIdx := range n.data {
		s := n.data[curIdx]
		if s.ID != id {
			newNode.data[idx] = s
			idx++
		}
	}
	newNode.inode = n.inode
	return newNode
}
