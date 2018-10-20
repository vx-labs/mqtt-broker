package state

import (
	"log"
	"time"

	"github.com/weaveworks/mesh"

	"github.com/vx-labs/mqtt-broker/set"
)

type state struct {
	set   set.LWW
	self  mesh.PeerName
	onAdd func(payload string)
	onDel func(payload string)
}

func New(self mesh.PeerName, onAdd, onDel func(string)) *state {
	s := &state{
		set:   set.NewLWW(),
		self:  self,
		onAdd: onAdd,
		onDel: onDel,
	}
	go s.startGC(6 * time.Hour)
	return s
}

func (st *state) startGC(maxage time.Duration) {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		now := time.Now()
		limit := now.UnixNano() - maxage.Nanoseconds()
		st.set = st.set.RemoveOlder(limit)
	}
}
func (st *state) Encode() [][]byte {
	return st.set.Serialize()
}

func (st *state) GossipData() mesh.GossipData {
	return &state{
		set:  st.set.Copy(),
		self: st.self,
	}
}
func (st *state) status() {
	log.Printf("state status")
	st.Iterate(func(key string, added, removed bool) error {
		log.Printf("  entry %x: added=%v removed=%v", key[0:8], added, removed)
		return nil
	})
}

func (st *state) merge(other *state) set.LWW {
	delta := st.set.Merge(other.set)
	addCount, delCount := 0, 0
	delta.Iterate(func(key string, entry set.Entry) error {
		if entry.IsAdded() {
			addCount++
			if st.onAdd != nil {
				st.onAdd(key)
			}
		}
		if entry.IsDeleted() {
			delCount++
			if st.onDel != nil {
				st.onDel(key)
			}
		}
		return nil
	})
	/*if addCount+delCount > 0 {
		log.Printf("INFO: synchronized state with mesh: %d entrie(s) were added, %d deleted", addCount, delCount)
	}*/
	return delta
}
func (st *state) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	otherState, ok := other.(*state)
	if !ok {
		panic("invalid state data received")
	}
	st.merge(otherState)
	return st
}

func (st *state) MergeDelta(buf []byte) (delta mesh.GossipData) {
	set, err := set.DecodeLWW([][]byte{buf})
	if err != nil {
		panic(err)
	}
	otherState := &state{
		set: set,
	}
	return &state{
		set: st.merge(otherState),
	}
}

func (st *state) Add(ev string) {
	st.set.Set(ev)
}

func (st *state) Remove(ev string) {
	st.set.Remove(ev)
}

func (st *state) Iterate(f func(ev string, added, deleted bool) error) error {
	return st.set.Iterate(func(ev string, entry set.Entry) error {
		return f(ev, entry.IsAdded(), entry.IsDeleted())
	})
}
