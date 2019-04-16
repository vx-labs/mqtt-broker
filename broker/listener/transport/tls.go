package transport

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/armon/go-proxyproto"

	"github.com/vx-labs/mqtt-broker/broker/listener"
)

type tlsListener struct {
	port     int
	listener net.Listener
}

type tlsTransport struct {
	ch    net.Conn
	state tls.ConnectionState
}

func (t *tlsTransport) Name() string {
	return "tls"
}

func (t *tlsTransport) Encrypted() bool {
	return true
}
func (t *tlsTransport) EncryptionState() *tls.ConnectionState {
	return &t.state
}
func (t *tlsTransport) RemoteAddress() string {
	return t.ch.RemoteAddr().String()
}
func (t *tlsTransport) Channel() listener.TimeoutReadWriteCloser {
	return t.ch
}
func (t *tlsTransport) Close() error {
	return t.ch.Close()
}

func NewTLSTransport(port int, TLSConfig *tls.Config, ch chan<- listener.Transport) (net.Listener, error) {
	listener := &tlsListener{
		port: port,
	}
	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", listener.port))
	if err != nil {
		return nil, err
	}
	proxyList := &proxyproto.Listener{Listener: tcp}

	l := tls.NewListener(proxyList, TLSConfig)
	listener.listener = l
	go listener.acceptLoop(ch)
	return l, nil
}

func (t *tlsListener) acceptLoop(ch chan<- listener.Transport) {
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
		t.queueSession(c, ch)
	}
}

func (t *tlsListener) queueSession(c *tls.Conn, ch chan<- listener.Transport) {
	ch <- &tlsTransport{
		ch:    c,
		state: c.ConnectionState(),
	}
}
