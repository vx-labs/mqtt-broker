package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/vx-labs/mqtt-broker/vaultacme"
	"go.uber.org/zap"
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
	mux.HandleFunc("/mqtt", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("INFO: starting websocket negociation with %s", r.RemoteAddr)
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Printf("ERR: websocket negociation with %s failed: %v", r.RemoteAddr, err)
			return
		}

		tlsConn := conn.(*tls.Conn)
		listener.queueSession(&Conn{
			conn:  conn,
			state: tlsConn.ConnectionState(),
		}, handler)
	})
	ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), TLSConfig)
	if err != nil {
		log.Fatalf("failed to start WSS listener: %v", err)
	}
	listener.listener = ln
	go http.Serve(ln, mux)
	return ln, nil
}

func (t *wssListener) queueSession(c *Conn, handler func(Metadata) error) {
	state := c.state
	handler(Metadata{
		Channel:         c,
		Encrypted:       true,
		EncryptionState: &state,
		Name:            "wss",
		RemoteAddress:   c.RemoteAddr().String(),
	})
}
