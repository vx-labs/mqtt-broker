package raft

import (
	"context"
	"errors"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	"github.com/hashicorp/raft"
	"github.com/vx-labs/mqtt-broker/adapters/cp/pb"
	"github.com/vx-labs/mqtt-broker/network"
	"go.uber.org/zap"
)

func (s *raftlayer) Serve() error {
	lis, err := s.rpcService.ListenTCP()
	if err != nil {
		return err
	}
	s.rpcListener = lis
	s.grpcServer = grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterLayerServer(s.grpcServer, s)
	grpc_prometheus.Register(s.grpcServer)
	go s.grpcServer.Serve(lis)
	return nil
}
func (s *raftlayer) SendEvent(ctx context.Context, input *pb.SendEventInput) (*pb.SendEventOutput, error) {
	return &pb.SendEventOutput{}, s.ApplyEvent(input.Payload)
}

func (s *raftlayer) PrepareShutdown(ctx context.Context, input *pb.PrepareShutdownInput) (*pb.PrepareShutdownOutput, error) {
	if s.raft.VerifyLeader().Error() != nil {
		return nil, errors.New("current node is not leader")
	}
	err := s.raft.RemoveServer(raft.ServerID(input.ID), input.Index, 0).Error()
	if err == nil {
		s.logger.Info("removed leaving node", zap.String("leaving_node", input.ID))
	}
	return &pb.PrepareShutdownOutput{}, err
}
func (s *raftlayer) RequestAdoption(ctx context.Context, input *pb.RequestAdoptionInput) (*pb.RequestAdoptionOutput, error) {
	if s.raft.VerifyLeader().Error() != nil {
		return nil, errors.New("current node is not leader")
	}
	s.syncMembers()
	err := s.raft.AddVoter(raft.ServerID(input.ID), raft.ServerAddress(input.Address), 0, 30*time.Second).Error()
	if err == nil {
		s.logger.Info("added new node", zap.String("new_node", input.ID))
	}
	return &pb.RequestAdoptionOutput{}, err
}
