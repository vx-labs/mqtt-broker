package rpc

import (
	"errors"
	"log"
	"sync"
)

var (
	ErrPoolNotFound = errors.New("pool not found")
)

type Caller struct {
	pools map[string]*Pool
	mutex sync.Mutex
}

func NewCaller() *Caller {
	return &Caller{
		pools: map[string]*Pool{},
	}
}
func (c *Caller) Call(addr string, job RPCJob) error {
	c.mutex.Lock()
	var err error
	data, ok := c.pools[addr]
	if !ok {
		log.Printf("INFO: creating a new RPC Pool targeting address %s", addr)
		data, err = NewPool(addr)
		if err != nil {
			c.mutex.Unlock()
			return err
		}
		c.pools[addr] = data
	}
	log.Printf("INFO: pool %s aquired", addr)
	c.mutex.Unlock()
	log.Printf("INFO: calling RPC targetting %s", addr)
	return data.Call(job)
}
func (c *Caller) Cancel(addr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	log.Printf("INFO: closing RPC pool toward %s", addr)
	pool, ok := c.pools[addr]
	if !ok {
		return ErrPoolNotFound
	}
	pool.Cancel()
	delete(c.pools, addr)
	return nil
}
