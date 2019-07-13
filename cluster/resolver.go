package cluster

import (
	"log"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"google.golang.org/grpc/resolver"
)

type Discoverer interface {
	EndpointsByService(name string) ([]*pb.NodeService, error)
	ByID(id string) (peers.Peer, error)
	On(event string, handler func(peers.Peer)) func()
}

func newResolver(d Discoverer) resolver.Builder {
	return &meshResolver{
		peers:         d,
		subscriptions: []func(){},
	}
}

type meshResolver struct {
	peers         Discoverer
	subscriptions []func()
}

func (r *meshResolver) Close() {
	for _, f := range r.subscriptions {
		f()
	}
}
func (r *meshResolver) Scheme() string {
	return "mesh"
}

func (r *meshResolver) updateConn(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) func(peer peers.Peer) {
	return func(p peers.Peer) {
		peers, err := r.peers.EndpointsByService(target.Endpoint)
		if err != nil {
			log.Printf("ERR: failed to search peers for service %s", target.Endpoint)
			return
		}
		addresses := make([]resolver.Address, len(peers))
		for idx, peer := range peers {
			addresses[idx] = resolver.Address{
				Addr:       peer.NetworkAddress,
				ServerName: peer.Peer,
				Type:       resolver.Backend,
				Metadata:   nil,
			}
		}
		cc.UpdateState(resolver.State{
			Addresses:     addresses,
			ServiceConfig: nil,
		})
	}
}
func (r *meshResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	cancelCreate := r.peers.On(peers.PeerCreated, r.updateConn(target, cc, opts))
	cancelDelete := r.peers.On(peers.PeerDeleted, r.updateConn(target, cc, opts))
	r.subscriptions = append(r.subscriptions, cancelCreate, cancelDelete)
	r.updateConn(target, cc, opts)
	return r, nil
}

func (r *meshResolver) ResolveNow(opts resolver.ResolveNowOption) {
}
