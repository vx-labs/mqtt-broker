package rpc

import (
	"errors"
	"log"
	"sync/atomic"
	"unsafe"

	iradix "github.com/hashicorp/go-immutable-radix"
)

var (
	ErrPoolNotFound = errors.New("pool not found")
)

type Caller struct {
	pools *iradix.Tree
}

func NewCaller() *Caller {
	return &Caller{
		pools: iradix.New(),
	}
}
func (c *Caller) cas(old, new *iradix.Tree) bool {
	oldPtr := (*unsafe.Pointer)(unsafe.Pointer(&c.pools))
	return atomic.CompareAndSwapPointer(oldPtr, unsafe.Pointer(old), unsafe.Pointer(new))
}
func (c *Caller) Call(addr string, job RPCJob) error {
	var pool (*Pool)
	data, ok := c.pools.Get([]byte(addr))
	if !ok {
		log.Printf("INFO: creating a new RPC Pool targeting address %s", addr)
		var err error
		pool, err = NewPool(addr)
		if err != nil {
			return err
		}
		for {
			old := c.pools
			new, _, _ := old.Insert([]byte(addr), pool)
			if c.cas(old, new) {
				break
			}
		}
	} else {
		pool = data.(*Pool)
	}
	return pool.Call(job)
}
func (c *Caller) Cancel(addr string) error {
	pool, ok := c.pools.Get([]byte(addr))
	if !ok {
		return ErrPoolNotFound
	}
	for {
		old := c.pools
		new, _, _ := old.Delete([]byte(addr))
		if c.cas(old, new) {
			pool.(*Pool).Cancel()
			return nil
		}
	}
}
