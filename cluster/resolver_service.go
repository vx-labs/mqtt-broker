package cluster

import (
	"log"
	"net/url"
	"strings"

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

func contains(needle string, slice []string) bool {
	for _, s := range slice {
		if s == needle {
			return true
		}
	}
	return false
}
func containsAll(needles []string, slice []string) bool {
	if needles == nil || len(needles) == 0 {
		return true
	}
	for _, needle := range needles {
		if !contains(needle, slice) {
			return false
		}
	}
	return true
}

func tagsFilter(endpoint string) (string, []*pb.ServiceTag) {
	targetURL, err := url.Parse("mesh:///" + endpoint)
	if err != nil {
		return endpoint, nil
	}
	tagFilter := pb.ParseFilter(targetURL.Query())
	return strings.TrimPrefix(targetURL.Path, "/"), tagFilter
}
func (r *meshResolver) updateConn(target resolver.Target, cc resolver.ClientConn) {
	service, tagFilter := tagsFilter(target.Endpoint)
	peers, err := r.peers.EndpointsByService(service)
	if err != nil {
		log.Printf("ERR: failed to search peers for service %s", service)
		return
	}
	addresses := make([]resolver.Address, 0)
	loggableAddresses := make([]string, 0)
	for idx := range peers {
		peer := peers[idx]
		if pb.MatchFilter(tagFilter, peer.Tags) {
			loggableAddresses = append(loggableAddresses, peer.Peer)
			addresses = append(addresses, resolver.Address{
				Addr:       peer.NetworkAddress,
				ServerName: peer.Peer,
				Type:       resolver.Backend,
				Metadata:   nil,
			})
		}
	}
	//	r.logger.Debug("updated mesh resolver targets", zap.Strings("targets", loggableAddresses), zap.String("grpc_endpoint", service))
	cc.UpdateState(resolver.State{
		Addresses:     addresses,
		ServiceConfig: nil,
	})
}
func (r *meshResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	cancelCreate := r.peers.On(peers.PeerCreated, func(p peers.Peer) { r.updateConn(target, cc) })
	cancelDelete := r.peers.On(peers.PeerDeleted, func(p peers.Peer) { r.updateConn(target, cc) })
	cancelUpdate := r.peers.On(peers.PeerUpdated, func(p peers.Peer) { r.updateConn(target, cc) })
	r.subscriptions = append(r.subscriptions, cancelCreate, cancelDelete, cancelUpdate)
	r.updateConn(target, cc)
	return r, nil
}

func (r *meshResolver) ResolveNow(opts resolver.ResolveNowOption) {
}
