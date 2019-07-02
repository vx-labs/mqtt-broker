package api

type clientCreateOpts struct {
	UseStaging bool
	Email      string
}

type Opt func(o *clientCreateOpts)

func WithEmail(addr string) Opt {
	return func(o *clientCreateOpts) {
		o.Email = addr
	}
}

func WithStagingAPI() Opt {
	return func(o *clientCreateOpts) {
		o.UseStaging = true
	}
}

func getOpts(opts []Opt) *clientCreateOpts {
	o := &clientCreateOpts{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
