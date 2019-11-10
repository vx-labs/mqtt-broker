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
	queues "github.com/vx-labs/mqtt-broker/queues/pb"

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
	Connect(context.Context, transport.Metadata, *packet.Connect) (string, string, *packet.ConnAck, error)
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

type QueuesStore interface {
	StreamMessages(ctx context.Context, id string, f func(uint64, *packet.Publish) error) error
	AckMessage(ctx context.Context, id string, ackOffset uint64) error
}

type endpoint struct {
	id         string
	mutex      sync.Mutex
	sessions   *btree.BTree
	queues     QueuesStore
	transports []net.Listener
	broker     Broker
	mesh       cluster.Mesh
	logger     *zap.Logger
}

func (local *endpoint) newSession(metadata transport.Metadata) error {
	metadata.Endpoint = local.id
	local.runLocalSession(metadata)
	return nil
}
func (local *endpoint) Close() error {
	for idx := range local.transports {
		local.transports[idx].Close()
	}
	return nil
}

type localSession struct {
	id        string
	token     string
	encoder   *encoder.Encoder
	logger    *zap.Logger
	transport io.Closer
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
	queuesConn, err := mesh.DialService("queues?raft_status=leader")
	if err != nil {
		panic(err)
	}
	local := &endpoint{
		broker:   brokerpb.NewClient(brokerConn),
		sessions: btree.New(2),
		queues:   queues.NewClient(queuesConn),
		id:       id,
		mesh:     mesh,
		logger:   logger,
	}
	if config.TCPPort > 0 {
		tcpTransport, err := transport.NewTCPTransport(config.TCPPort, local.newSession)
		if err != nil {
			local.logger.Warn("failed to start listener",
				zap.String("transport", "tcp"), zap.Error(err))
		} else {
			local.logger.Info("started listener",
				zap.String("transport", "tcp"))
			local.transports = append(local.transports, tcpTransport)
		}
	}
	if config.WSPort > 0 {
		wsTransport, err := transport.NewWSTransport(config.WSPort, local.newSession)
		if err != nil {
			local.logger.Warn("failed to start listener",
				zap.String("transport", "ws"), zap.Error(err))
		} else {
			local.logger.Info("started listener",
				zap.String("transport", "ws"))
			local.transports = append(local.transports, wsTransport)
		}
	}
	if config.WSSPort > 0 {
		wssTransport, err := transport.NewWSSTransport(ctx, config.TLSCommonName, config.WSSPort, logger, local.newSession)
		if err != nil {
			local.logger.Warn("failed to start listener",
				zap.String("transport", "wss"), zap.Error(err))
		} else {
			local.logger.Info("started listener",
				zap.String("transport", "wss"))
			local.transports = append(local.transports, wssTransport)
		}
	}
	if config.TLSPort > 0 {
		tlsTransport, err := transport.NewTLSTransport(ctx, config.TLSCommonName, config.TLSPort, logger, local.newSession)
		if err != nil {
			local.logger.Warn("failed to start listener",
				zap.String("transport", "tls"), zap.Error(err))
		} else {
			local.logger.Info("started listener",
				zap.String("transport", "tls"))
			local.transports = append(local.transports, tlsTransport)
		}

	}
	return local
}
