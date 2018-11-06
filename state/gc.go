package state

import (
	"time"
)

func (s *Store) gCRunner() {
	for range time.Tick(1 * time.Hour) {
		s.gC(time.Now().Truncate(8 * time.Hour).UnixNano())
	}
}
func (s *Store) gC(limit int64) {
	set := s.backend.Dump()
	set.Range(func(idx int, entry Entry) {
		if isEntryRemoved(entry) {
			if entry.GetLastDeleted() < limit {
				s.backend.DeleteEntry(entry)
			}
		}
	})
}
