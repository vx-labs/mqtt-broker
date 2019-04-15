package rpc

import (
	"fmt"
	"io"
	"log"
	"net"

	sessions "github.com/vx-labs/mqtt-broker/sessions"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/broker/rpc --go_out=plugins=grpc:. rpc.proto

type broker interface {
	ListSessions() (sessions.SessionList, error)
	CloseSession(id string) error
	DistributeMessage(*MessagePublished) error
}

type server struct {
	broker   broker
	listener io.Closer
	server   *grpc.Server
}

func New(port int, handler broker) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("WARN: failed to start rpc listener: %v", err)
		return nil
	}
	s := grpc.NewServer()
	server := &server{
		broker:   handler,
		listener: lis,
		server:   s,
	}
	RegisterBrokerServiceServer(s, server)
	go s.Serve(lis)
	log.Printf("INFO: started RPC listener on %s", lis.Addr().String())
	return lis
}
func (s *server) Close() error {
	s.server.Stop()
	return s.listener.Close()
}
func (s *server) CloseSession(ctx context.Context, input *CloseSessionInput) (*CloseSessionOutput, error) {
	return &CloseSessionOutput{ID: input.ID}, s.broker.CloseSession(input.ID)
}
func (s *server) ListSessions(ctx context.Context, filters *SessionFilter) (*ListSessionsOutput, error) {
	set, err := s.broker.ListSessions()
	if err != nil {
		return nil, err
	}
	return &ListSessionsOutput{Sessions: set.Sessions}, nil
}
func (s *server) DistributeMessage(ctx context.Context, msg *MessagePublished) (*MessagePublishedOutput, error) {
	err := s.broker.DistributeMessage(msg)
	return &MessagePublishedOutput{}, err
}
