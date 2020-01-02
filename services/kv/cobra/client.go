package cobra

import (
	"log"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"

	"github.com/vx-labs/mqtt-broker/services/kv/pb"
)

func getClient(adapter discovery.DiscoveryAdapter) *pb.Client {
	conn, err := adapter.DialService("kv")
	if err != nil {
		log.Fatalf("failed to connect to kv: %v", err)
	}
	return pb.NewClient(conn)
}
