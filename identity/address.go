package identity

import "fmt"

// Address represents a host and port combinaison
type Address interface {
	Port() int
	Host() string
	String() string
}
type address struct {
	port    int
	host    string
	address string
}

func (a *address) String() string {
	if a.address == "" {
		a.address = fmt.Sprintf("%s:%d", a.host, a.port)
	}
	return a.address
}
func (a *address) Host() string {
	return a.host
}

func (a *address) Port() int {
	return a.port
}
