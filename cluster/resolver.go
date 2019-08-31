package cluster

import (
	"log"

	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

type Discoverer interface {
	EndpointsByService(name string) ([]*pb.NodeService, error)
	ByID(id string) (peers.Peer, error)
	On(event string, handler func(peers.Peer)) func()
}

func newResolver(d Discoverer, logger *zap.Logger) resolver.Builder {
	return &meshResolver{
		peers:         d,
		logger:        logger.With(zap.String("emitter", "grpc_mesh_resolver")),
		subscriptions: []func(){},
	}
}

type meshResolver struct {
	peers         Discoverer
	logger        *zap.Logger
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

func (r *meshResolver) updateConn(target resolver.Target, cc resolver.ClientConn) {
	peers, err := r.peers.EndpointsByService(target.Endpoint)
	if err != nil {
		log.Printf("ERR: failed to search peers for service %s", target.Endpoint)
		return
	}
	addresses := make([]resolver.Address, len(peers))
	loggableAddresses := make([]string, len(peers))
	for idx, peer := range peers {
		loggableAddresses[idx] = peer.Peer
		addresses[idx] = resolver.Address{
			Addr:       peer.NetworkAddress,
			ServerName: peer.Peer,
			Type:       resolver.Backend,
			Metadata:   nil,
		}
	}
	r.logger.Info("updated mesh resolver targets", zap.Strings("targets", loggableAddresses), zap.String("grpc_endpoint", target.Endpoint))
	cc.UpdateState(resolver.State{
		Addresses:     addresses,
		ServiceConfig: nil,
	})
}
func (r *meshResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	cancelCreate := r.peers.On(peers.PeerCreated, func(p peers.Peer) { r.updateConn(target, cc) })
	cancelDelete := r.peers.On(peers.PeerDeleted, func(p peers.Peer) { r.updateConn(target, cc) })
	r.subscriptions = append(r.subscriptions, cancelCreate, cancelDelete)
	r.updateConn(target, cc)
	return r, nil
}

func (r *meshResolver) ResolveNow(opts resolver.ResolveNowOption) {
}
