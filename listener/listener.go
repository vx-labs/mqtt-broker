package listener

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/google/btree"
	brokerpb "github.com/vx-labs/mqtt-broker/broker/pb"
	publishQueue "github.com/vx-labs/mqtt-broker/queues/publish"

	"github.com/vx-labs/mqtt-broker/cluster"
	"github.com/vx-labs/mqtt-broker/transport"
	"github.com/vx-labs/mqtt-protocol/encoder"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

var (
	ErrSessionNotFound = errors.New("session not found on this endpoint")
)

type Broker interface {
	Connect(context.Context, transport.Metadata, *packet.Connect) (string, *packet.ConnAck, error)
	Disconnect(context.Context, string, *packet.Disconnect) error
	Publish(context.Context, string, *packet.Publish) (*packet.PubAck, error)
	Subscribe(context.Context, string, *packet.Subscribe) (*packet.SubAck, error)
	Unsubscribe(context.Context, string, *packet.Unsubscribe) (*packet.UnsubAck, error)
	CloseSession(context.Context, string) error
	PingReq(context.Context, string, *packet.PingReq) (*packet.PingResp, error)
}

type Endpoint interface {
	Publish(ctx context.Context, id string, publish *packet.Publish) error
	CloseSession(context.Context, string) error
	Close() error
}

type endpoint struct {
	id         string
	mutex      sync.Mutex
	sessions   *btree.BTree
	transports []net.Listener
	broker     Broker
	mesh       cluster.Mesh
	logger     *zap.Logger
}

func (local *endpoint) newSession(metadata transport.Metadata) error {
	metadata.Endpoint = local.id
	go local.runLocalSession(metadata)
	return nil
}
func (local *endpoint) Close() error {
	for idx := range local.transports {
		local.transports[idx].Close()
	}
	return nil
}
func (local *endpoint) CloseSession(ctx context.Context, id string) error {
	local.mutex.Lock()
	defer local.mutex.Unlock()
	session := local.sessions.Delete(&localSession{
		id: id,
	})
	if session != nil {
		return session.(*localSession).transport.Close()
	}
	return ErrSessionNotFound
}
func (local *endpoint) Publish(ctx context.Context, id string, publish *packet.Publish) error {
	session := local.sessions.Get(&localSession{
		id: id,
	})
	if session == nil {
		local.logger.Warn("session not found", zap.String("node_id", local.id), zap.String("session_id", id))
		return ErrSessionNotFound
	}
	session.(*localSession).queue.Enqueue(&publishQueue.Message{Publish: publish})
	return nil
}

type localSession struct {
	id        string
	encoder   *encoder.Encoder
	queue     publishQueue.Queue
	transport io.Closer
	quit      chan struct{}
}

func (local *localSession) Less(remote btree.Item) bool {
	return strings.Compare(local.id, remote.(*localSession).id) > 0
}

type Config struct {
	TCPPort       int
	TLSCommonName string
	TLSPort       int
	WSSPort       int
	WSPort        int
}

func New(id string, logger *zap.Logger, mesh cluster.Mesh, config Config) *endpoint {
	ctx := context.Background()
	brokerConn, err := mesh.DialService("broker")
	if err != nil {
		panic(err)
	}
	brokerClient := brokerpb.NewClient(brokerConn)
	local := &endpoint{
		broker:   brokerClient,
		sessions: btree.New(2),
		id:       id,
		mesh:     mesh,
		logger:   logger,
	}
	if config.TCPPort > 0 {
		tcpTransport, err := transport.NewTCPTransport(config.TCPPort, local.newSession)
		if err != nil {
			local.logger.Warn("failed to start listener", zap.String("node_id", id),
				zap.String("transport", "tcp"), zap.Error(err))
		} else {
			local.logger.Info("started listener", zap.String("node_id", id),
				zap.String("transport", "tcp"))
			local.transports = append(local.transports, tcpTransport)
		}
	}
	if config.WSPort > 0 {
		wsTransport, err := transport.NewWSTransport(config.WSPort, local.newSession)
		if err != nil {
			local.logger.Warn("failed to start listener", zap.String("node_id", id),
				zap.String("transport", "ws"), zap.Error(err))
		} else {
			local.logger.Info("started listener", zap.String("node_id", id),
				zap.String("transport", "ws"))
			local.transports = append(local.transports, wsTransport)
		}
	}
	if config.WSSPort > 0 {
		wssTransport, err := transport.NewWSSTransport(ctx, config.TLSCommonName, config.WSSPort, logger, local.newSession)
		if err != nil {
			local.logger.Warn("failed to start listener", zap.String("node_id", id),
				zap.String("transport", "wss"), zap.Error(err))
		} else {
			local.logger.Info("started listener", zap.String("node_id", id),
				zap.String("transport", "wss"))
			local.transports = append(local.transports, wssTransport)
		}
	}
	if config.TLSPort > 0 {
		tlsTransport, err := transport.NewTLSTransport(ctx, config.TLSCommonName, config.TLSPort, logger, local.newSession)
		if err != nil {
			local.logger.Warn("failed to start listener", zap.String("node_id", id),
				zap.String("transport", "tls"), zap.Error(err))
		} else {
			local.logger.Info("started listener", zap.String("node_id", id),
				zap.String("transport", "tls"))
			local.transports = append(local.transports, tlsTransport)
		}

	}
	return local
}
