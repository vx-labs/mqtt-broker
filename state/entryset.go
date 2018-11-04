package state

import "github.com/golang/protobuf/proto"

type EntrySet interface {
	proto.Message
	Range(func(idx int, entry Entry))
	AtIndex(int) Entry
	Set(int, Entry)
	Length() int
	New() EntrySet
	Append(Entry)
}
