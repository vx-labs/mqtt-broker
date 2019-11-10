package kv

import (
	fmt "fmt"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/kv/pb"
	"github.com/vx-labs/mqtt-broker/network"

	grpc "google.golang.org/grpc"

	"go.uber.org/zap"
)

func (b *server) Shutdown() {
	for _, lis := range b.listeners {
		lis.Close()
	}
	b.store.Close()
}
func (b *server) JoinServiceLayer(name string, logger *zap.Logger, config cluster.ServiceConfig, rpcConfig cluster.ServiceConfig, mesh cluster.DiscoveryLayer) {
	b.state = cluster.NewRaftServiceLayer(name, logger, config, rpcConfig, mesh)
	go func() {
		err := b.state.Start(name, b)
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		var lastDeadline uint64 = 0
		for range ticker.C {
			if b.state.IsLeader() {
				keys, err := b.store.ListExpiredKeys(lastDeadline, uint64(time.Now().UnixNano()))
				if err != nil {
					b.logger.Error("failed to list expired messages", zap.Error(err))
					continue
				}
				if len(keys) > 0 {
					err := b.commitEvent(&pb.KVStateTransition{
						Event: &pb.KVStateTransition_DeleteBatch{
							DeleteBatch: &pb.KVStateTransitionValueBatchDeleted{
								KeyMDs: keys,
							},
						},
					})
					if err != nil {
						b.logger.Error("failed to delete expired keys", zap.Int("expired_key_count", len(keys)), zap.Error(err))
					} else {
						b.logger.Info("deleted expired keys", zap.Int("expired_key_count", len(keys)))
						lastDeadline = keys[len(keys)-1].Deadline
					}
				}
			}
		}
	}()
}
func (m *server) Health() string {
	return m.state.Health()
}
func (m *server) Serve(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}
	s := grpc.NewServer(
		network.GRPCServerOptions()...,
	)
	pb.RegisterKVServiceServer(s, m)
	grpc_prometheus.Register(s)
	go s.Serve(lis)
	return lis
}
