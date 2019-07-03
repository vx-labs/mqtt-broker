package listener

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/vx-labs/mqtt-broker/broker/listener/pb"

	"google.golang.org/grpc"
)

type server struct {
	endpoint Endpoint
}

func (s *server) Publish(ctx context.Context, input *pb.PublishInput) (*pb.PublishOutput, error) {
	return &pb.PublishOutput{}, s.endpoint.Publish(input.ID, input.Publish)
}
func (s *server) CloseSession(ctx context.Context, input *pb.CloseSessionInput) (*pb.CloseSessionOutput, error) {
	return &pb.CloseSessionOutput{}, s.endpoint.CloseSession(input.ID)
}
func (s *server) Close(ctx context.Context, input *pb.CloseInput) (*pb.CloseOutput, error) {
	return &pb.CloseOutput{}, s.endpoint.Close()
}

func Serve(local Endpoint, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterListenerServiceServer(grpcServer, &server{endpoint: local})
	return grpcServer.Serve(lis)
}
