package cobra

import (
	"log"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"

	"github.com/vx-labs/mqtt-broker/services/queues/pb"
)

func getClient(adapter discovery.DiscoveryAdapter) *pb.Client {
	conn, err := adapter.DialService("queues", "rpc")
	if err != nil {
		log.Fatalf("failed to connect to messages: %v", err)
	}
	return pb.NewClient(conn)
}
