package rpc

import grpc "google.golang.org/grpc"

type RPCJob func(BrokerServiceClient) error
type Pool struct {
	jobs chan chan RPCJob
	quit chan struct{}
}

func (a *Pool) Call(job RPCJob) error {
	worker := <-a.jobs
	worker <- job
	return nil
}
func (a *Pool) Cancel() {
	close(a.quit)
}
func NewPool(addr string) (*Pool, error) {
	c := &Pool{
		jobs: make(chan chan RPCJob),
		quit: make(chan struct{}),
	}
	jobs := make(chan RPCJob)
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := NewBrokerServiceClient(conn)
	for i := 0; i < 5; i++ {
		go func() {
			for {
				select {
				case <-c.quit:
					return
				case c.jobs <- jobs:
				}

				select {
				case <-c.quit:
					return
				case job := <-jobs:
					job(client)
				}
			}
		}()
	}
	return c, nil
}
