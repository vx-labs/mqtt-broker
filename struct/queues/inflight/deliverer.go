package inflight

import (
	"time"

	"github.com/vx-labs/mqtt-protocol/packet"
)

func startDeliverer(mid int32, jobs chan chan Job, sender func(*packet.Publish) error) *acknowleger {
	ch := make(chan *packet.PubAck)
	cancel := make(chan struct{})
	quit := make(chan struct{})
	job := make(chan Job)
	go func() {
		defer close(ch)
		defer close(job)
		defer close(quit)
		for {
			select {
			case <-cancel:
				return
			case jobs <- job:
			}
			select {
			case <-cancel:
				return
			case j := <-job:
				j.publish.MessageId = mid
			loop:
				for {
					err := sender(j.publish)
					if err != nil {
						break loop
					}
					if j.publish.Header.Qos == 0 {
						break loop
					}
					ticker := time.NewTicker(10 * time.Second)
					select {
					case <-ch:
						ticker.Stop()
						j.onAck()
						break loop
					case <-ticker.C:
						j.publish.Header.Dup = true
						continue
					case <-cancel:
						ticker.Stop()
						return
					}
				}
			}

		}
	}()
	return &acknowleger{
		cancel: cancel,
		ch:     ch,
		mid:    mid,
		quit:   quit,
	}
}
