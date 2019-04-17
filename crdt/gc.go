package crdt

import (
	"io"
	"time"
)

func ExpireAfter8Hours() int64 {
	return time.Now().Truncate(8 * time.Hour).UnixNano()
}

func GCEntries(limit int64, next func() (Entry, error), gc func(id string) error) error {
	for {
		entry, err := next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if IsEntryRemoved(entry) {
			if entry.GetLastDeleted() < limit {
				err = gc(entry.GetID())
				if err != nil {
					return err
				}
			}
		}
	}
}
