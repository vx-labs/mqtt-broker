package rpc

import (
	grpc "google.golang.org/grpc"
)

type RPCJob func(BrokerServiceClient) error
type Pool struct {
	jobs chan chan JobWrap
	quit chan struct{}
}

type JobWrap struct {
	job  RPCJob
	done chan error
}

func (a *Pool) Call(job RPCJob) error {
	worker := <-a.jobs
	ch := make(chan error)
	worker <- JobWrap{
		job:  job,
		done: ch,
	}
	return <-ch
}
func (a *Pool) Cancel() {
	close(a.quit)
}
func NewPool(addr string) (*Pool, error) {
	c := &Pool{
		jobs: make(chan chan JobWrap),
		quit: make(chan struct{}),
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	go func() {
		<-c.quit
		conn.Close()
	}()

	client := NewBrokerServiceClient(conn)
	jobs := make(chan JobWrap)
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
				job.done <- job.job(client)
				close(job.done)
			}
		}
	}()
	return c, nil
}
