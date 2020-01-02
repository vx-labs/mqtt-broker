package cobra

import (
	"log"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"

	"github.com/vx-labs/mqtt-broker/services/auth/pb"
)

func getClient(adapter discovery.DiscoveryAdapter) *pb.Client {
	conn, err := adapter.DialService("auth")
	if err != nil {
		log.Fatalf("failed to connect to subscriptions: %v", err)
	}
	return pb.NewClient(conn)
}
