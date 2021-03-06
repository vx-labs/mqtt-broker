package consul

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

func newConsulResolver(d *api.Client, logger *zap.Logger) resolver.Builder {
	return &meshResolver{
		peers:  d,
		logger: logger.With(zap.String("emitter", "consul_mesh_resolver")),
	}
}

type meshResolver struct {
	peers  *api.Client
	logger *zap.Logger
}

func (r *meshResolver) Close() {
}
func (r *meshResolver) Scheme() string {
	return "consul"
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

func tagsFilter(endpoint string) (string, string) {
	targetURL, err := url.Parse("consul:///" + endpoint)
	if err != nil {
		return endpoint, ""
	}
	tagFilter := targetURL.Query().Get("tag")
	return strings.TrimPrefix(targetURL.Path, "/"), tagFilter
}
func (r *meshResolver) updateConn(target resolver.Target, cc resolver.ClientConn) {
	service, tagFilter := tagsFilter(target.Endpoint)
	peers, _, err := r.peers.Health().Service(service, tagFilter, true, nil)
	if err != nil {
		r.logger.Warn("failed to search consul for service", zap.Error(err))
		return
	}
	addresses := make([]resolver.Address, 0)
	loggableAddresses := make([]string, 0)
	for idx := range peers {
		peer := peers[idx]
		loggableAddresses = append(loggableAddresses, peer.Service.ID)
		addresses = append(addresses, resolver.Address{
			Addr:       fmt.Sprintf("%s:%d", peer.Service.Address, peer.Service.Port),
			ServerName: peer.Service.Meta["node_id"],
			Type:       resolver.Backend,
			Metadata:   nil,
		})
	}
	r.logger.Debug("updated mesh resolver targets", zap.Strings("targets", loggableAddresses), zap.String("grpc_endpoint", service))
	cc.UpdateState(resolver.State{
		Addresses:     addresses,
		ServiceConfig: nil,
	})
}
func (r *meshResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	service, tag := tagsFilter(target.Endpoint)
	go func() {
		var idx uint64 = 0
		for {
			_, meta, err := r.peers.Health().Service(service, tag, true, &api.QueryOptions{
				WaitIndex: idx,
				WaitTime:  60 * time.Second,
			})
			if err != nil {
				r.logger.Warn("failed to search consul for service", zap.Error(err))
				<-time.After(1 * time.Second)
				continue
			}
			if idx == meta.LastIndex {
				continue
			}
			r.updateConn(target, cc)
			idx = meta.LastIndex
		}
	}()
	return r, nil
}

func (r *meshResolver) ResolveNow(opts resolver.ResolveNowOption) {
}
