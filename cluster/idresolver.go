package cluster

import (
	"log"
	"strings"

	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"google.golang.org/grpc/resolver"
)

func newIDResolver(d Discoverer) resolver.Builder {
	return &meshIDResolver{
		peers:         d,
		subscriptions: []func(){},
	}
}

type meshIDResolver struct {
	peers         Discoverer
	subscriptions []func()
}

func (r *meshIDResolver) Close() {
	for _, f := range r.subscriptions {
		f()
	}
}

// meshid://service+id/
func (r *meshIDResolver) Scheme() string {
	return "meshid"
}

func (r *meshIDResolver) updateConn(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) func(peer peers.Peer) {
	return func(_ peers.Peer) {
		tokens := strings.Split(target.Endpoint, "+")
		service := tokens[0]
		peerID := tokens[1]
		peer, err := r.peers.EndpointsByService(service)
		if err != nil {
			log.Printf("ERR: failed to search peer for id %s", target.Endpoint)
			cc.UpdateState(resolver.State{
				Addresses:     []resolver.Address{},
				ServiceConfig: nil,
			})
			return
		}
		addresses := make([]resolver.Address, 1)

		for _, service := range peer {
			if service.Peer == peerID {
				addresses[0] = resolver.Address{
					Addr:       service.NetworkAddress,
					ServerName: peerID,
					Type:       resolver.Backend,
					Metadata:   nil,
				}
				break
			}
		}
		cc.UpdateState(resolver.State{
			Addresses:     addresses,
			ServiceConfig: nil,
		})
	}
}
func (r *meshIDResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	cancelCreate := r.peers.On(peers.PeerCreated, r.updateConn(target, cc, opts))
	cancelDelete := r.peers.On(peers.PeerDeleted, r.updateConn(target, cc, opts))
	r.subscriptions = append(r.subscriptions, cancelCreate, cancelDelete)
	r.updateConn(target, cc, opts)
	return r, nil
}

func (r *meshIDResolver) ResolveNow(opts resolver.ResolveNowOption) {
}
