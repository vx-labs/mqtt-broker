package cobra

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/network"
	"github.com/vx-labs/mqtt-broker/services/subscriptions/pb"
	"google.golang.org/grpc"
)

func getClient(config *viper.Viper) *pb.Client {
	host := config.GetString("host")
	opts := network.GRPCClientOptions()
	conn, err := grpc.Dial(host,
		append(opts, grpc.WithTimeout(800*time.Millisecond))...)
	if err != nil {
		log.Fatalf("failed to connect %s: %v", host, err)
	}
	return pb.NewClient(conn)
}
