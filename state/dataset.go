package state

import (
	"github.com/golang/protobuf/proto"
	"github.com/weaveworks/mesh"
)

var _ mesh.GossipData = &Dataset{}

type Dataset struct {
	backend EntrySet
}

func (d *Dataset) Encode() [][]byte {
	payload, err := proto.Marshal(d.backend)
	if err != nil {
		return nil
	}
	return [][]byte{payload}
}
func (set Dataset) Merge(data mesh.GossipData) mesh.GossipData {
	remoteSet := data.(*Dataset)
	idx := make(map[string]int, set.backend.Length())
	set.backend.Range(func(i int, e Entry) {
		idx[e.GetID()] = i
	})
	delta := remoteSet.backend.New()

	remoteSet.backend.Range(func(_ int, remote Entry) {
		local, ok := idx[remote.GetID()]
		if !ok {
			// Session not found in our store, add it to delta
			delta.Append(remote)
			set.backend.Append(remote)
		} else {
			if isEntryOutdated(set.backend.AtIndex(local), remote) {
				delta.Append(remote)
				set.backend.Set(local, remote)
			}
		}
	})
	return &Dataset{
		backend: delta,
	}
}
