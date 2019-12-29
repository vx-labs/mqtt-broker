package mesh

import (
	"runtime"
	"time"

	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
)

func memUsage() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

func (b *MeshDiscoveryAdapter) oSStatsReporter() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		m := memUsage()
		nbRoutines := runtime.NumGoroutine()
		nbCores := runtime.NumCPU()
		err := b.peers.Update(b.id, func(self pb.Peer) pb.Peer {
			self.ComputeUsage = &pb.ComputeUsage{
				Cores:      int64(nbCores),
				Goroutines: int64(nbRoutines),
			}
			self.MemoryUsage = &pb.MemoryUsage{
				Alloc:      m.Alloc,
				TotalAlloc: m.TotalAlloc,
				NumGC:      m.NumGC,
				Sys:        m.Sys,
			}
			return self
		})
		if err == nil {
			b.syncMeta()
		}
		<-ticker.C
	}
}
