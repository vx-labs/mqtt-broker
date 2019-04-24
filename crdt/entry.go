package crdt

// An Entry is a CRDT struct
type Entry interface {
	GetID() string
	GetLastAdded() int64
	GetLastDeleted() int64
}

func IsEntryAdded(s Entry) bool {
	return s.GetLastAdded() > 0 && s.GetLastAdded() > s.GetLastDeleted()
}
func IsEntryRemoved(s Entry) bool {
	return s.GetLastDeleted() > 0 && s.GetLastAdded() < s.GetLastDeleted()
}

func GetLastEntryUpdate(s Entry) int64 {
	if s.GetLastAdded() > s.GetLastDeleted() {
		return s.GetLastAdded()
	}
	return s.GetLastDeleted()
}

func IsEntryOutdated(s Entry, remote Entry) (outdated bool) {
	return GetLastEntryUpdate(s) < GetLastEntryUpdate(remote)
}
