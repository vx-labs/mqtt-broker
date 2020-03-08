package cobra

import (
	"log"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"

	"github.com/vx-labs/mqtt-broker/services/sessions/pb"
)

func getClient(adapter discovery.DiscoveryAdapter) *pb.Client {
	conn, err := adapter.DialService("sessions", "rpc")
	if err != nil {
		log.Fatalf("failed to connect to sessions: %v", err)
	}
	return pb.NewClient(conn)
}
