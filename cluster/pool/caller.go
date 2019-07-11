package pool

import (
	"errors"
	"strings"
	"sync"

	"github.com/google/btree"
)

var (
	ErrPoolNotFound = errors.New("pool not found")
)

type Caller struct {
	pools *btree.BTree
	mutex sync.Mutex
}

func (p *Pool) Less(remote btree.Item) bool {
	return strings.Compare(p.address, remote.(*Pool).address) == -1
}

func NewCaller() *Caller {
	return &Caller{
		pools: btree.New(2),
	}
}
func (c *Caller) Call(addr string, job RPCJob) error {
	var pool *Pool
	data := c.pools.Get(&Pool{address: addr})
	if data == nil {
		c.mutex.Lock()
		data = c.pools.Get(&Pool{address: addr})
		if data == nil {
			newpool, err := NewPool(addr)
			if err != nil {
				c.mutex.Unlock()
				return err
			}
			old := c.pools.ReplaceOrInsert(newpool)
			if old != nil {
				old.(*Pool).Cancel()
			}
			data = newpool
		}
		c.mutex.Unlock()
	}
	pool = data.(*Pool)
	return pool.Call(job)
}
func (c *Caller) Cancel(addr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	pool := c.pools.Get(&Pool{address: addr})
	if pool == nil {
		return ErrPoolNotFound
	}
	pool.(*Pool).Cancel()
	c.pools.Delete(pool)
	return nil
}
