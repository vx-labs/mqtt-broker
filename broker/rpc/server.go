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

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/broker/rpc --go_out=plugins=grpc:. broker.proto

type broker interface {
	ListSessions() ([]*sessions.Session, error)
}

type server struct {
	broker   broker
	listener io.Closer
	server   *grpc.Server
}

func New(port int, handler broker) io.Closer {
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
	RegisterBrokerServer(s, server)
	go s.Serve(lis)
	log.Printf("INFO: started RPC listener on port %d", port)
	return server
}
func (s *server) Close() error {
	s.server.Stop()
	return s.listener.Close()
}
func (s *server) ListSessions(ctx context.Context, filters *SessionFilter) (*ListSessionsOutput, error) {
	set, err := s.broker.ListSessions()
	if err != nil {
		return nil, err
	}
	return &ListSessionsOutput{Sessions: set}, nil
}
