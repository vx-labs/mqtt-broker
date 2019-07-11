package pool

type Job func() error
type Pool struct {
	jobs chan chan Job
	quit chan struct{}
}

func (a *Pool) Call(job Job) error {
	worker := <-a.jobs
	worker <- job
	return nil
}
func (a *Pool) Cancel() {
	close(a.quit)
}
func NewPool(count int) *Pool {
	c := &Pool{
		jobs: make(chan chan Job),
	}

	jobs := make(chan Job)
	for i := 0; i < count; i++ {
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
					job()
				}
			}
		}()
	}
	return c
}
