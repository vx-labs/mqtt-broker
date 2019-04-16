package state

import (
	"github.com/golang/protobuf/proto"
)

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
