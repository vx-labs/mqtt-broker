package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/vx-labs/mqtt-broker/vaultacme"
	"go.uber.org/zap"
)

type tlsListener struct {
	listener net.Listener
}

func NewTLSTransport(ctx context.Context, cn string, port int, logger *zap.Logger, handler func(Metadata) error) (net.Listener, error) {
	listener := &tlsListener{}
	TLSConfig, err := vaultacme.GetConfig(ctx, cn, logger)
	if err != nil {
		return nil, err
	}

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	proxyList := &proxyproto.Listener{Listener: tcp}

	l := tls.NewListener(proxyList, TLSConfig)
	listener.listener = l
	go listener.acceptLoop(handler)
	return l, nil
}

func (t *tlsListener) acceptLoop(handler func(Metadata) error) {
	var tempDelay time.Duration
	for {
		rawConn, err := t.listener.Accept()

		if err != nil {
			if err.Error() == fmt.Sprintf("accept tcp %v: use of closed network connection", t.listener.Addr()) {
				return
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Printf("WARN: accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			log.Printf("ERROR: connection handling failed: %v", err)
			t.listener.Close()
			return
		}
		c, ok := rawConn.(*tls.Conn)
		if !ok {
			c.Close()
			continue
		}
		err = c.Handshake()
		if err != nil {
			log.Printf("ERROR: tls handshake failed: %v", err)
			c.Close()
			continue
		}
		t.queueSession(c, handler)
	}
}

func (t *tlsListener) queueSession(c *tls.Conn, handler func(Metadata) error) {
	state := c.ConnectionState()
	go handler(Metadata{
		Channel:         c,
		Encrypted:       true,
		EncryptionState: &state,
		Name:            "tls",
		RemoteAddress:   c.RemoteAddr().String(),
	})
}
