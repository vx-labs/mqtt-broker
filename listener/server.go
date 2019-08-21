package listener

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/listener/pb"
	"github.com/vx-labs/mqtt-broker/network"

	"google.golang.org/grpc"
)

type server struct {
	endpoint Endpoint
}

func (s *server) SendPublish(ctx context.Context, input *pb.SendPublishInput) (*pb.SendPublishOutput, error) {
	err := s.endpoint.Publish(ctx, input.ID, input.Publish)
	if err == ErrSessionNotFound {
		return &pb.SendPublishOutput{}, status.Error(codes.NotFound, err.Error())
	}
	return &pb.SendPublishOutput{}, err
}
func (s *server) SendBatchPublish(ctx context.Context, input *pb.SendBatchPublishInput) (*pb.SendBatchPublishOutput, error) {
	for _, recipient := range input.ID {
		s.endpoint.Publish(ctx, recipient, input.Publish)
	}
	return &pb.SendBatchPublishOutput{}, nil
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
		network.GRPCServerOptions()...,
	)
	pb.RegisterListenerServiceServer(grpcServer, &server{endpoint: local})
	grpc_prometheus.Register(grpcServer)
	go grpcServer.Serve(lis)
	return lis
}
