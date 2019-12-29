package identity

import "sync"

type Identity interface {
	Name() string
	AdvertisedAddress() string
	AdvertisedPort() int
	BindAddress() string
	BindPort() int
}

type Catalog interface {
	Register(i Identity)
	Get(name string) Identity
}

type catalog struct {
	mtx  sync.Mutex
	data map[string]Identity
}

func NewCatalog() Catalog {
	return &catalog{
		data: map[string]Identity{},
	}
}

func (c *catalog) Register(i Identity) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.data[i.Name()] = i
}
func (c *catalog) Get(name string) Identity {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.data[name]
}
