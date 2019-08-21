package pool

import (
	"fmt"

	"github.com/vx-labs/mqtt-broker/network"
	grpc "google.golang.org/grpc"
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
		network.GRPCClientOptions()...,
	)
	if err != nil {
		return nil, err
	}
	c.conn = conn
	return c, nil
}
