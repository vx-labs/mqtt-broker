package broker

import (
	"fmt"
	"io"
	"log"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/vx-labs/mqtt-broker/broker/pb"
	sessions "github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/transport"
	packet "github.com/vx-labs/mqtt-protocol/packet"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type broker interface {
	ListSessions() (sessions.SessionSet, error)
	CloseSession(ctx context.Context, id string) error
	DistributeMessage(*pb.MessagePublished) error
	Connect(context.Context, transport.Metadata, *packet.Connect) (string, *packet.ConnAck, error)
	Disconnect(context.Context, string, *packet.Disconnect) error
	Publish(context.Context, string, *packet.Publish) (*packet.PubAck, error)
	Subscribe(context.Context, string, *packet.Subscribe) (*packet.SubAck, error)
	Unsubscribe(context.Context, string, *packet.Unsubscribe) (*packet.UnsubAck, error)
	PingReq(context.Context, string, *packet.PingReq) (*packet.PingResp, error)
}

type server struct {
	broker   broker
	listener io.Closer
	server   *grpc.Server
}

func Serve(port int, handler broker) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("WARN: failed to start rpc listener: %v", err)
		return nil
	}
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	server := &server{
		broker:   handler,
		listener: lis,
		server:   s,
	}
	pb.RegisterBrokerServiceServer(s, server)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	return lis
}
func (s *server) Close() error {
	s.server.Stop()
	return s.listener.Close()
}
func (s *server) CloseSession(ctx context.Context, input *pb.CloseSessionInput) (*pb.CloseSessionOutput, error) {
	return &pb.CloseSessionOutput{ID: input.ID}, s.broker.CloseSession(ctx, input.ID)
}
func (s *server) ListSessions(ctx context.Context, filters *pb.SessionFilter) (*pb.ListSessionsOutput, error) {
	set, err := s.broker.ListSessions()
	if err != nil {
		return nil, err
	}
	out := []*sessions.Metadata{}
	for _, session := range set {
		out = append(out, &session.Metadata)
	}
	return &pb.ListSessionsOutput{Sessions: out}, nil
}
func (s *server) DistributeMessage(ctx context.Context, msg *pb.MessagePublished) (*pb.MessagePublishedOutput, error) {
	err := s.broker.DistributeMessage(msg)
	return &pb.MessagePublishedOutput{}, err
}

func (s *server) Connect(ctx context.Context, input *pb.ConnectInput) (*pb.ConnectOutput, error) {
	id, connack, err := s.broker.Connect(ctx, transport.Metadata{
		Encrypted:     input.TransportMetadata.Encrypted,
		Name:          input.TransportMetadata.Name,
		RemoteAddress: input.TransportMetadata.RemoteAddress,
		Endpoint:      input.TransportMetadata.Endpoint,
	}, input.Connect)
	return &pb.ConnectOutput{
		ID:      id,
		ConnAck: connack,
	}, err
}

func (s *server) Disconnect(ctx context.Context, input *pb.DisconnectInput) (*pb.DisconnectOutput, error) {
	err := s.broker.Disconnect(ctx, input.ID, input.Disconnect)
	return &pb.DisconnectOutput{}, err
}
func (s *server) Publish(ctx context.Context, input *pb.PublishInput) (*pb.PublishOutput, error) {
	puback, err := s.broker.Publish(ctx, input.ID, input.Publish)
	return &pb.PublishOutput{PubAck: puback}, err
}
func (s *server) Subscribe(ctx context.Context, input *pb.SubscribeInput) (*pb.SubscribeOutput, error) {
	suback, err := s.broker.Subscribe(ctx, input.ID, input.Subscribe)
	return &pb.SubscribeOutput{SubAck: suback}, err
}
func (s *server) Unsubscribe(ctx context.Context, input *pb.UnsubscribeInput) (*pb.UnsubscribeOutput, error) {
	unsuback, err := s.broker.Unsubscribe(ctx, input.ID, input.Unsubscribe)
	return &pb.UnsubscribeOutput{UnsubAck: unsuback}, err
}
func (s *server) PingReq(ctx context.Context, input *pb.PingReqInput) (*pb.PingReqOutput, error) {
	pingresp, err := s.broker.PingReq(ctx, input.ID, input.PingReq)
	return &pb.PingReqOutput{PingResp: pingresp}, err
}
