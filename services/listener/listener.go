package listener

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/stream"
	"github.com/vx-labs/mqtt-broker/struct/queues/inflight"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/btree"
	brokerpb "github.com/vx-labs/mqtt-broker/services/broker/pb"
	queues "github.com/vx-labs/mqtt-broker/services/queues/pb"

	kv "github.com/vx-labs/mqtt-broker/services/kv/pb"
	messages "github.com/vx-labs/mqtt-broker/services/messages/pb"
	"github.com/vx-labs/mqtt-broker/transport"
	"github.com/vx-labs/mqtt-protocol/encoder"
	"github.com/vx-labs/mqtt-protocol/packet"
	"go.uber.org/zap"
)

func SigningKey() string {
	return os.Getenv("JWT_SIGN_KEY")
}

type Token struct {
	SessionID     string `json:"session_id"`
	SessionTenant string `json:"session_tenant"`
	jwt.StandardClaims
}

func DecodeSessionToken(signKey string, signedToken string) (Token, error) {
	token, err := jwt.ParseWithClaims(signedToken, &Token{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(signKey), nil
	})
	if err != nil {
		return Token{}, err
	}
	if claims, ok := token.Claims.(*Token); ok && token.Valid {
		return *claims, nil
	}
	return Token{}, err
}

var (
	ErrSessionNotFound = errors.New("session not found on this endpoint")
)

type Broker interface {
	Connect(context.Context, transport.Metadata, *packet.Connect) (string, string, string, *packet.ConnAck, error)
	Disconnect(context.Context, string, *packet.Disconnect) error
	Publish(context.Context, string, *packet.Publish) (*packet.PubAck, error)
	Subscribe(context.Context, string, *packet.Subscribe) (*packet.SubAck, error)
	Unsubscribe(context.Context, string, *packet.Unsubscribe) (*packet.UnsubAck, error)
	CloseSession(context.Context, string) error
	PingReq(context.Context, string, *packet.PingReq) (string, *packet.PingResp, error)
}

type Endpoint interface {
	Publish(ctx context.Context, id string, publish *packet.Publish) error
	CloseSession(context.Context, string) error
	Close() error
}

type QueuesStore interface {
	StreamMessages(ctx context.Context, id string, f func(uint64, *packet.Publish) error) error
	AckMessage(ctx context.Context, id string, ackOffset uint64) error
	GetQueueStatistics(ctx context.Context, queueID string) (*queues.QueueStatistics, error)
}

type endpoint struct {
	id           string
	mutex        sync.Mutex
	sessions     *btree.BTree
	queues       QueuesStore
	transports   []net.Listener
	broker       Broker
	messages     *messages.Client
	streamClient *stream.Client
	kv           *kv.Client
	logger       *zap.Logger
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
	id           string
	token        string
	refreshToken string
	cancel       context.CancelFunc
	encoder      *encoder.Encoder
	timer        int32
	inflights    *inflight.Queue
	logger       *zap.Logger
	transport    transport.Metadata
	poller       sync.Once
	pollerCh     chan error
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

func New(id string, logger *zap.Logger, mesh discovery.DiscoveryAdapter, config Config) *endpoint {
	ctx := context.Background()
	brokerConn, err := mesh.DialService("broker", "rpc")
	if err != nil {
		panic(err)
	}
	queuesConn, err := mesh.DialService("queues", "rpc")
	if err != nil {
		panic(err)
	}
	messagesConn, err := mesh.DialService("messages", "rpc")
	if err != nil {
		panic(err)
	}
	kvConn, err := mesh.DialService("kv", "rpc")
	if err != nil {
		panic(err)
	}
	local := &endpoint{
		broker:   brokerpb.NewClient(brokerConn),
		sessions: btree.New(2),
		queues:   queues.NewClient(queuesConn),
		messages: messages.NewClient(messagesConn),
		kv:       kv.NewClient(kvConn),
		id:       id,
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
