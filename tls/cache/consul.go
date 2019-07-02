package cache

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
)

type ConsulLocker struct {
	consul *consul.Client
}

func (c *ConsulLocker) Lock(ctx context.Context, prefix string) (*consul.Lock, error) {
	l, err := c.consul.LockKey(fmt.Sprintf("%s/lock", prefix))
	if err != nil {
		return nil, err
	}
	_, err = l.Lock(ctx.Done())
	if err != nil {
		return nil, err
	}
	return l, err
}
func (c *ConsulLocker) Unlock(ctx context.Context) error {
	return nil
}

func NewConsulLocker(api *consul.Client) *ConsulLocker {
	return &ConsulLocker{
		consul: api,
	}
}
