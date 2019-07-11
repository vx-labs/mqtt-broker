package inflight

import (
	"time"

	"github.com/vx-labs/mqtt-protocol/packet"
)

func startDeliverer(mid int32, jobs chan chan *packet.Publish, sender func(*packet.Publish) error) *acknowleger {
	ch := make(chan *packet.PubAck)
	cancel := make(chan struct{})
	quit := make(chan struct{})
	ticker := time.NewTicker(10 * time.Second)
	job := make(chan *packet.Publish)
	go func() {
		defer ticker.Stop()
		defer close(ch)
		defer close(job)
		defer close(quit)
		for {
			select {
			case <-cancel:
				return
			case jobs <- job:
			case publish := <-job:
				publish.MessageId = mid
				err := sender(publish)
				if err != nil {
					continue
				}
				if publish.Header.Qos == 0 {
					continue
				}
				select {
				case <-ch:
					continue
				case <-ticker.C:
					publish.Header.Dup = true
					continue
				case <-cancel:
					return
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
