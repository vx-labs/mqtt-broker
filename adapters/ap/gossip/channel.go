package gossip

import (
	"github.com/hashicorp/memberlist"
)

type simpleBroadcast []byte

func (b simpleBroadcast) Message() []byte                       { return []byte(b) }
func (b simpleBroadcast) Invalidates(memberlist.Broadcast) bool { return false }
func (b simpleBroadcast) Finished()                             {}

type simpleFullStateBroadcast []byte

func (b simpleFullStateBroadcast) Message() []byte                       { return []byte(b) }
func (b simpleFullStateBroadcast) Invalidates(memberlist.Broadcast) bool { return true }
func (b simpleFullStateBroadcast) Finished()                             {}
