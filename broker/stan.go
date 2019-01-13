package broker

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	stan "github.com/nats-io/go-nats-streaming"
)

type STANMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Tenant    string    `json:"tenant"`
	Payload   []byte    `json:"payload"`
	Topic     []byte    `json:"topic"`
}

func exportToSTAN(config Config, ch chan STANMessage) error {
	id := uuid.New().String()
	sc, err := stan.Connect("events", id,
		stan.NatsURL(config.NATSURL),
		stan.Pings(10, 5),
	)
	if err != nil {
		return err
	}
	go func() {
		for msg := range ch {
			payload, err := json.Marshal(msg)
			if err != nil {
				log.Printf("WARN: failed to publish message to STAN: %v", err)
			} else {
				sc.Publish("vx.iot.mqtt.message_published", payload)
			}
		}
	}()

	return nil
}
