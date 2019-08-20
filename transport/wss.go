package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/vx-labs/mqtt-broker/vaultacme"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

type wssListener struct {
	listener net.Listener
}

func NewWSSTransport(ctx context.Context, cn string, port int, logger *zap.Logger, handler func(Metadata) error) (net.Listener, error) {
	listener := &wssListener{}
	TLSConfig, err := vaultacme.GetConfig(ctx, cn, logger)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	server := &websocket.Server{
		Handshake: func(config *websocket.Config, r *http.Request) error {
			return nil
		},
		Handler: websocket.Handler(func(conn *websocket.Conn) {
			listener.queueSession(conn, handler)
		}),
		Config: websocket.Config{
			Version: websocket.ProtocolVersionHybi13,
		},
	}
	mux.Handle("/mqtt", server)
	ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), TLSConfig)
	if err != nil {
		logger.Fatal("failed to start WSS listener", zap.Error(err))
	}
	listener.listener = ln
	go http.Serve(ln, mux)
	return ln, nil
}

func (t *wssListener) queueSession(c *websocket.Conn, handler func(Metadata) error) {
	r := c.Request()
	handler(Metadata{
		Channel:         c,
		Encrypted:       true,
		EncryptionState: nil,
		Name:            "wss",
		RemoteAddress:   r.RemoteAddr,
	})
}
