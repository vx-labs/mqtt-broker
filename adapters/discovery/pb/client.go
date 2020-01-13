package pb

import (
	context "context"
	fmt "fmt"
	"log"
	"strings"
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

func (d *DiscoveryAdapter) Members() ([]*Peer, error) {
	out, err := d.api.ListMembers(d.ctx, &ListMembersInput{})
	if err != nil {
		return nil, err
	}
	return out.Peers, nil
}
func (d *DiscoveryAdapter) EndpointsByService(name string) ([]*NodeService, error) {
	out, err := d.api.GetEndpoints(d.ctx, &GetEndpointsInput{ServiceName: name})
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
func (d *DiscoveryAdapter) RegisterTCPService(id, name, address string) error {
	_, err := d.api.RegisterService(d.ctx, &RegisterServiceInput{ServiceID: id, ServiceName: name, NetworkAddress: address})
	return err
}
func (d *DiscoveryAdapter) RegisterUDPService(id, name, address string) error {
	_, err := d.api.RegisterService(d.ctx, &RegisterServiceInput{ServiceID: id, ServiceName: name, NetworkAddress: address})
	return err
}
func (d *DiscoveryAdapter) RegisterGRPCService(id, name, address string) error {
	_, err := d.api.RegisterService(d.ctx, &RegisterServiceInput{ServiceID: id, ServiceName: name, NetworkAddress: address})
	return err
}
func (d *DiscoveryAdapter) UnregisterService(id string) error {
	_, err := d.api.UnregisterService(d.ctx, &UnregisterServiceInput{ServiceID: id})
	return err
}
func (d *DiscoveryAdapter) AddServiceTag(service, key, value string) error {
	_, err := d.api.AddServiceTag(d.ctx, &AddServiceTagInput{ServiceID: service, TagKey: key, TagValue: value})
	return err
}
func (d *DiscoveryAdapter) RemoveServiceTag(service, key string) error {
	_, err := d.api.RemoveServiceTag(d.ctx, &RemoveServiceTagInput{ServiceID: service, TagKey: key})
	return err
}
func (d *DiscoveryAdapter) DialService(name string, tags ...string) (*grpc.ClientConn, error) {
	key := fmt.Sprintf("pb:///%s", name)
	if len(tags) > 0 {
		key = fmt.Sprintf("%s?%s", key, strings.Join(tags, "&"))
	}
	return grpc.Dial(key,
		network.GRPCClientOptions()...,
	)
}
func (d *DiscoveryAdapter) Shutdown() error {
	return d.conn.Close()
}
