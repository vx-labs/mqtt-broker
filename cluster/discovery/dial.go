package discovery

import (
	fmt "fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

func (m *discoveryLayer) DialService(name string) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("mesh:///%s", name),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithInsecure(), grpc.WithAuthority(name), grpc.WithBalancerName(roundrobin.Name))
}
func (m *discoveryLayer) DialAddress(service, id string, f func(*grpc.ClientConn) error) error {
	return m.rpcCaller.Call(fmt.Sprintf("%s+%s", service, id), f)
}
