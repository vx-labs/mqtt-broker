package pb

import (
	context "context"
	"errors"
	fmt "fmt"
	"log"
	"net"
	"time"

	"github.com/vx-labs/mqtt-broker/network"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

type DiscoveryAdapter struct {
	ctx  context.Context
	id   string
	conn *grpc.ClientConn
	api  DiscoveryServiceClient
}

func NewPBDiscoveryAdapter(ctx context.Context, nodeid, host string, logger *zap.Logger) *DiscoveryAdapter {
	opts := network.GRPCClientOptions()
	conn, err := grpc.Dial(host,
		append(opts, grpc.WithTimeout(800*time.Millisecond))...)
	if err != nil {
		log.Fatalf("failed to connect %s: %v", host, err)
	}
	c := &DiscoveryAdapter{
		ctx:  ctx,
		conn: conn,
		id:   nodeid,
		api:  NewDiscoveryServiceClient(conn),
	}
	resolver.Register(NewPBResolver(c, logger))
	return c
}

func (d *DiscoveryAdapter) EndpointsByService(name, tag string) ([]*NodeService, error) {
	out, err := d.api.GetEndpoints(d.ctx, &GetEndpointsInput{ServiceName: name, Tag: tag})
	if err != nil {
		return nil, err
	}
	return out.NodeServices, nil
}
func (d *DiscoveryAdapter) streamEndpoints(ctx context.Context, name string) (chan []*NodeService, error) {
	ch := make(chan []*NodeService)
	stream, err := d.api.StreamEndpoints(ctx, &GetEndpointsInput{ServiceName: name})
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(ch)
		for {
			next, err := stream.Recv()
			if err != nil {
				return
			}
			ch <- next.NodeServices
		}
	}()
	return ch, nil
}
func (d *DiscoveryAdapter) ListenTCP(id, name string, port int, advertizedAddress string) (net.Listener, error) {
	return nil, errors.New("unsupported")
}
func (d *DiscoveryAdapter) ListenUDP(id, name string, port int, advertizedAddress string) (net.PacketConn, error) {
	return nil, errors.New("unsupported")
}
func (d *DiscoveryAdapter) DialService(name string, tag string) (*grpc.ClientConn, error) {
	key := fmt.Sprintf("pb:///%s", name)
	key = fmt.Sprintf("%s?tag=%s", key, tag)

	return grpc.Dial(key,
		network.GRPCClientOptions()...,
	)
}
func (d *DiscoveryAdapter) Shutdown() error {
	return d.conn.Close()
}
