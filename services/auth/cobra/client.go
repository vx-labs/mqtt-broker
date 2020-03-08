package cobra

import (
	"log"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"

	"github.com/vx-labs/mqtt-broker/services/auth/pb"
)

func getClient(adapter discovery.DiscoveryAdapter) *pb.Client {
	conn, err := adapter.DialService("auth", "rpc")
	if err != nil {
		log.Fatalf("failed to connect to auth: %v", err)
	}
	return pb.NewClient(conn)
}
