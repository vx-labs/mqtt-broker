package cluster

import (
	"runtime"
	"time"
)

func memUsage() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}
func (b *memberlistMesh) oSStatsReporter() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		m := memUsage()
		nbRoutines := runtime.NumGoroutine()
		nbCores := runtime.NumCPU()
		self, err := b.peers.ByID(b.id)
		if err != nil {
			return
		}
		self.ComputeUsage = &ComputeUsage{
			Cores:      int64(nbCores),
			Goroutines: int64(nbRoutines),
		}
		self.MemoryUsage = &MemoryUsage{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			NumGC:      m.NumGC,
			Sys:        m.Sys,
		}
		b.peers.Upsert(self)
		<-ticker.C
	}
}
