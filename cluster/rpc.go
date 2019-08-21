package cluster

import (
	fmt "fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"google.golang.org/grpc"
)

func ServeRPC(port int, l pb.LayerServer) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterLayerServer(s, l)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	return lis
}
