package cobra

import (
	"log"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"

	"github.com/vx-labs/mqtt-broker/services/subscriptions/pb"
)

func getClient(adapter discovery.DiscoveryAdapter) *pb.Client {
	conn, err := adapter.DialService("subscriptions", "rpc")
	if err != nil {
		log.Fatalf("failed to connect to subscriptions: %v", err)
	}
	return pb.NewClient(conn)
}
