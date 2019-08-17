package cluster

import (
	fmt "fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"google.golang.org/grpc"
)

func ServeRPC(port int, l pb.LayerServer) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	pb.RegisterLayerServer(s, l)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	return lis
}
