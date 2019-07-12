package cluster

import (
	proto "github.com/golang/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
)

type channel struct {
	key   string
	bcast *memberlist.TransmitLimitedQueue
}

// We use a simple broadcast implementation in which items are never invalidated by others.
type simpleBroadcast []byte

func (b simpleBroadcast) Message() []byte                       { return []byte(b) }
func (b simpleBroadcast) Invalidates(memberlist.Broadcast) bool { return false }
func (b simpleBroadcast) Finished()                             {}

// Broadcast enqueues a message for broadcasting.
func (c *channel) Broadcast(b []byte) {
	b, err := proto.Marshal(&pb.Part{Key: c.key, Data: b})
	if err != nil {
		return
	}
	c.bcast.QueueBroadcast(simpleBroadcast(b))
}
