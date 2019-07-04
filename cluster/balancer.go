package cluster

import (
	"context"
	"errors"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type failoverPickerBuilder struct {
}

type failoverPicker struct {
	subConns  []balancer.SubConn
	addresses []string
}

func (p *failoverPicker) Pick(ctx context.Context, opts balancer.PickOptions) (balancer.SubConn, func(balancer.DoneInfo), error) {
	if len(p.subConns) < 1 {
		return nil, nil, errors.New("no connection available")
	}
	sc := p.subConns[0]
	return sc, nil, nil
}

func (f *failoverPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	var scs []balancer.SubConn
	picker := &failoverPicker{
		subConns: scs,
	}
	for address, sc := range readySCs {
		picker.subConns = append(picker.subConns, sc)
		picker.addresses = append(picker.addresses, address.Addr)
	}
	return picker
}

func init() {
	balancer.Register(base.NewBalancerBuilderWithConfig("failover", &failoverPickerBuilder{}, base.Config{HealthCheck: true}))
}
