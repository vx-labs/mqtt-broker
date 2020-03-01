package nomad

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

func newNomadResolver(d *api.Client, logger *zap.Logger) resolver.Builder {
	return &meshResolver{
		peers:  d,
		logger: logger.With(zap.String("emitter", "nomad_mesh_resolver")),
	}
}

type meshResolver struct {
	peers  *api.Client
	logger *zap.Logger
}

func (r *meshResolver) Close() {
}
func (r *meshResolver) Scheme() string {
	return "nomad"
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
	targetURL, err := url.Parse("nomad:///" + endpoint)
	if err != nil {
		return endpoint, ""
	}
	tagFilter := targetURL.Query().Get("tag")
	return strings.TrimPrefix(targetURL.Path, "/"), tagFilter
}
func (r *meshResolver) updateConn(target resolver.Target, cc resolver.ClientConn, idx uint64) (uint64, error) {
	service, tagFilter := tagsFilter(target.Endpoint)
	peers, meta, err := r.peers.Health().Service(service, tagFilter, true, &api.QueryOptions{
		WaitIndex: idx,
		WaitTime:  60 * time.Second,
	})
	if err != nil {
		r.logger.Warn("failed to search nomad for service", zap.Error(err))
		return 0, err
	}
	addresses := make([]resolver.Address, 0)
	loggableAddresses := make([]string, 0)
	for idx := range peers {
		peer := peers[idx]
		nomadAllocID := strings.TrimPrefix(peer.Service.ID, "_nomad-task-")[0:36]

		loggableAddresses = append(loggableAddresses, peer.Service.ID)
		addresses = append(addresses, resolver.Address{
			Addr:       fmt.Sprintf("%s:%d", peer.Service.Address, peer.Service.Port),
			ServerName: nomadAllocID,
			Type:       resolver.Backend,
			Metadata:   nil,
		})
	}
	r.logger.Debug("updated mesh resolver targets", zap.Strings("targets", loggableAddresses), zap.String("grpc_endpoint", service))
	cc.UpdateState(resolver.State{
		Addresses:     addresses,
		ServiceConfig: nil,
	})
	return meta.LastIndex, nil
}
func (r *meshResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	go func() {
		var idx uint64 = 0
		for {
			newIdx, err := r.updateConn(target, cc, idx)
			if err != nil {
				r.logger.Error("failed to search for peers", zap.Error(err))
			} else {
				idx = newIdx
			}
		}
	}()
	return r, nil
}

func (r *meshResolver) ResolveNow(opts resolver.ResolveNowOption) {
}
