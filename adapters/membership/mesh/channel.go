package mesh

import (
	proto "github.com/golang/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/adapters/membership/pb"
)

type channel struct {
	key   string
	bcast *memberlist.TransmitLimitedQueue
}

type simpleBroadcast []byte

func (b simpleBroadcast) Message() []byte                       { return []byte(b) }
func (b simpleBroadcast) Invalidates(memberlist.Broadcast) bool { return false }
func (b simpleBroadcast) Finished()                             {}

type simpleFullStateBroadcast []byte

func (b simpleFullStateBroadcast) Message() []byte                       { return []byte(b) }
func (b simpleFullStateBroadcast) Invalidates(memberlist.Broadcast) bool { return true }
func (b simpleFullStateBroadcast) Finished()                             {}

// Broadcast enqueues a message for broadcasting.
func (c *channel) Broadcast(b []byte) {
	b, err := proto.Marshal(&pb.Part{Key: c.key, Data: b})
	if err != nil {
		return
	}
	c.bcast.QueueBroadcast(simpleBroadcast(b))
}
func (c *channel) BroadcastFullState(b []byte) {
	b, err := proto.Marshal(&pb.Part{Key: c.key, Data: b})
	if err != nil {
		return
	}
	c.bcast.QueueBroadcast(simpleFullStateBroadcast(b))
}
