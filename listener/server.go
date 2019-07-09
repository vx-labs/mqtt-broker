package listener

import (
	"context"
	"fmt"
	"log"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/listener/pb"

	"google.golang.org/grpc"
)

type server struct {
	endpoint Endpoint
}

func (s *server) SendPublish(ctx context.Context, input *pb.SendPublishInput) (*pb.SendPublishOutput, error) {
	return &pb.SendPublishOutput{}, s.endpoint.Publish(ctx, input.ID, input.Publish)
}
func (s *server) Shutdown(ctx context.Context, input *pb.ShutdownInput) (*pb.ShutdownOutput, error) {
	return &pb.ShutdownOutput{}, s.endpoint.CloseSession(ctx, input.ID)
}
func (s *server) Close(ctx context.Context, input *pb.CloseInput) (*pb.CloseOutput, error) {
	return &pb.CloseOutput{}, s.endpoint.Close()
}

func Serve(local Endpoint, port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	pb.RegisterListenerServiceServer(grpcServer, &server{endpoint: local})
	grpc_prometheus.Register(grpcServer)
	go grpcServer.Serve(lis)
	return lis
}
