package cobra

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/messages/pb"
	"google.golang.org/grpc"
)

func getClient(config *viper.Viper) *pb.Client {
	host := config.GetString("host")
	conn, err := grpc.Dial(host,
		grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(300*time.Millisecond))
	if err != nil {
		log.Fatalf("failed to connect %s: %v", host, err)
	}
	return pb.NewClient(conn)
}