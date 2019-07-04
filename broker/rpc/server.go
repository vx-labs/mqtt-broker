package rpc

import (
	"fmt"
	"io"
	"log"
	"net"

	sessions "github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/transport"
	packet "github.com/vx-labs/mqtt-protocol/packet"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/broker/rpc --go_out=plugins=grpc:. rpc.proto

type broker interface {
	ListSessions() (sessions.SessionSet, error)
	CloseSession(ctx context.Context, id string) error
	DistributeMessage(*MessagePublished) error
	Connect(context.Context, transport.Metadata, *packet.Connect) (string, *packet.ConnAck, error)
	Disconnect(context.Context, string, *packet.Disconnect) error
	Publish(context.Context, string, *packet.Publish) (*packet.PubAck, error)
	Subscribe(context.Context, string, *packet.Subscribe) (*packet.SubAck, error)
	Unsubscribe(context.Context, string, *packet.Unsubscribe) (*packet.UnsubAck, error)
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
	return &CloseSessionOutput{ID: input.ID}, s.broker.CloseSession(ctx, input.ID)
}
func (s *server) ListSessions(ctx context.Context, filters *SessionFilter) (*ListSessionsOutput, error) {
	set, err := s.broker.ListSessions()
	if err != nil {
		return nil, err
	}
	out := []*sessions.Metadata{}
	for _, session := range set {
		out = append(out, &session.Metadata)
	}
	return &ListSessionsOutput{Sessions: out}, nil
}
func (s *server) DistributeMessage(ctx context.Context, msg *MessagePublished) (*MessagePublishedOutput, error) {
	err := s.broker.DistributeMessage(msg)
	return &MessagePublishedOutput{}, err
}

func (s *server) Connect(ctx context.Context, input *ConnectInput) (*ConnectOutput, error) {
	id, connack, err := s.broker.Connect(ctx, transport.Metadata{
		Encrypted:     input.TransportMetadata.Encrypted,
		Name:          input.TransportMetadata.Name,
		RemoteAddress: input.TransportMetadata.RemoteAddress,
	}, input.Connect)
	return &ConnectOutput{
		ID:      id,
		ConnAck: connack,
	}, err
}

func (s *server) Disconnect(ctx context.Context, input *DisconnectInput) (*DisconnectOutput, error) {
	err := s.broker.Disconnect(ctx, input.ID, input.Disconnect)
	return &DisconnectOutput{}, err
}
func (s *server) Publish(ctx context.Context, input *PublishInput) (*PublishOutput, error) {
	puback, err := s.broker.Publish(ctx, input.ID, input.Publish)
	return &PublishOutput{PubAck: puback}, err
}
func (s *server) Subscribe(ctx context.Context, input *SubscribeInput) (*SubscribeOutput, error) {
	suback, err := s.broker.Subscribe(ctx, input.ID, input.Subscribe)
	return &SubscribeOutput{SubAck: suback}, err
}
func (s *server) Unsubscribe(ctx context.Context, input *UnsubscribeInput) (*UnsubscribeOutput, error) {
	unsuback, err := s.broker.Unsubscribe(ctx, input.ID, input.Unsubscribe)
	return &UnsubscribeOutput{UnsubAck: unsuback}, err
}
