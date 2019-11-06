package discovery

import (
	"runtime"
	"time"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
)

func memUsage() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}
func (b *discoveryLayer) keepaliveSender() {
	ticker := time.NewTicker(15 * time.Second)
	for {
		now := time.Now().UnixNano()
		b.peers.Update(b.id, func(self peers.Peer) peers.Peer {
			if self.LastContact < now {
				self.LastContact = now
			}
			return self
		})
		<-ticker.C
	}
}
func (b *discoveryLayer) deadNodeDeleter() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		deadline := time.Now().AddDate(0, 0, -1).UnixNano()
		peers, err := b.peers.All()
		if err == nil {
			for _, peer := range peers {
				if peer.LastContact < deadline {
					b.peers.Delete(peer.ID)
				}
			}
		}
		<-ticker.C
	}
}
func (b *discoveryLayer) oSStatsReporter() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		m := memUsage()
		nbRoutines := runtime.NumGoroutine()
		nbCores := runtime.NumCPU()
		err := b.Peers().Update(b.id, func(self peers.Peer) peers.Peer {
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
		if err != nil {
			return
		}
		<-ticker.C
	}
}
