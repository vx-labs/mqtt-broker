package crdt

import (
	"io"
	"time"
)

type ExpirationPolicy func() int64

func ExpireAfter8Hours() ExpirationPolicy {
	return func() int64 {
		return time.Now().Truncate(8 * time.Hour).UnixNano()
	}
}

func GCEntries(policy ExpirationPolicy, next func() (Entry, error), gc func(id string) error) error {
	limit := policy()
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
