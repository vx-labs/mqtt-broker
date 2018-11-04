package state

import "github.com/golang/protobuf/proto"

type Entry interface {
	proto.Message
	GetLastAdded() int64
	GetLastDeleted() int64
	GetID() string
}

func isEntryAdded(s Entry) bool {
	return s.GetLastAdded() > 0 && s.GetLastAdded() > s.GetLastDeleted()
}
func isEntryRemoved(s Entry) bool {
	return s.GetLastDeleted() > 0 && s.GetLastAdded() < s.GetLastDeleted()
}

func getLastEntryUpdate(s Entry) int64 {
	if s.GetLastAdded() > s.GetLastDeleted() {
		return s.GetLastAdded()
	}
	return s.GetLastDeleted()
}

func isEntryOutdated(s Entry, remote Entry) (outdated bool) {
	return getLastEntryUpdate(s) < getLastEntryUpdate(remote)
}
