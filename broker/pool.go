package broker

type Job func() error
type Pool struct {
	jobs chan chan JobWrap
	quit chan struct{}
}

type JobWrap struct {
	job  Job
	done chan error
}

func (a *Pool) Call(job Job) error {
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
func NewPool(count int) *Pool {
	c := &Pool{
		jobs: make(chan chan JobWrap),
	}

	jobs := make(chan JobWrap)
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
					job.done <- job.job()
					close(job.done)
				}
			}
		}()
	}
	return c
}
