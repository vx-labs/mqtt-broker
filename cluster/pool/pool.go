package pool

import (
	"fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type RPCJob func(*grpc.ClientConn) error
type Pool struct {
	address string
	conn    *grpc.ClientConn
}

type JobWrap struct {
	job  RPCJob
	done chan error
}

func (a *Pool) Call(job RPCJob) error {
	return job(a.conn)
}
func (a *Pool) Cancel() {
	a.conn.Close()
}
func NewPool(addr string) (*Pool, error) {
	c := &Pool{
		address: addr,
	}
	conn, err := grpc.Dial(fmt.Sprintf("meshid:///%s", addr),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithInsecure(), grpc.WithAuthority(addr),
		grpc.WithBalancerName(roundrobin.Name),
	)
	if err != nil {
		return nil, err
	}
	c.conn = conn
	return c, nil
}
