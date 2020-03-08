package pb

import (
	context "context"
	"log"
	"net/url"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

func NewPBResolver(d *DiscoveryAdapter, logger *zap.Logger) resolver.Builder {
	return &pbResolver{
		peers:  d,
		logger: logger.With(zap.String("emitter", "grpc_pb_resolver")),
	}
}

type pbResolver struct {
	peers  *DiscoveryAdapter
	logger *zap.Logger
}

type pbResolverSession struct {
	peers  *DiscoveryAdapter
	logger *zap.Logger
	cancel context.CancelFunc
}

func (r *pbResolverSession) Close() {
	r.cancel()
}
func (r *pbResolver) Scheme() string {
	return "pb"
}

func tagsFilter(endpoint string) (string, string) {
	targetURL, err := url.Parse("pb:///" + endpoint)
	if err != nil {
		return endpoint, ""
	}
	tagFilter := targetURL.Query().Get("tag")
	return strings.TrimPrefix(targetURL.Path, "/"), tagFilter
}
func (r *pbResolverSession) updateConn(target resolver.Target, cc resolver.ClientConn) {
	service, tagFilter := tagsFilter(target.Endpoint)
	peers, err := r.peers.EndpointsByService(service, tagFilter)
	if err != nil {
		log.Printf("ERR: failed to search peers for service %s: %v", service, err)
		return
	}
	addresses := make([]resolver.Address, 0)
	loggableAddresses := make([]string, 0)
	for idx := range peers {
		peer := peers[idx]
		if tagFilter == peer.Tag {
			loggableAddresses = append(loggableAddresses, peer.Peer)
			addresses = append(addresses, resolver.Address{
				Addr:       peer.NetworkAddress,
				ServerName: peer.Peer,
				Type:       resolver.Backend,
				Metadata:   nil,
			})
		}
	}
	r.logger.Debug("updated mesh resolver targets", zap.Strings("targets", loggableAddresses), zap.String("grpc_endpoint", service))
	cc.UpdateState(resolver.State{
		Addresses:     addresses,
		ServiceConfig: nil,
	})
}
func (r *pbResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(context.Background())
	session := &pbResolverSession{
		cancel: cancel,
		peers:  r.peers,
		logger: r.logger,
	}
	service, _ := tagsFilter(target.Endpoint)
	ch, err := r.peers.streamEndpoints(ctx, service)
	if err != nil {
		return nil, err
	}
	go func() {
		for range ch {
			session.updateConn(target, cc)
		}
		session.logger.Debug("stream closed")
	}()
	session.updateConn(target, cc)
	return session, nil
}

func (r *pbResolverSession) ResolveNow(opts resolver.ResolveNowOption) {
}
