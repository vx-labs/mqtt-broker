package transport

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	proxyproto "github.com/armon/go-proxyproto"
	"github.com/vx-labs/mqtt-broker/broker/listener"
)

type tcp struct {
	port     int
	listener net.Listener
}

type tcpTransport struct {
	ch net.Conn
}

func (t *tcpTransport) Name() string {
	return "tcp"
}

func (t *tcpTransport) Encrypted() bool {
	return false
}
func (t *tcpTransport) EncryptionState() *tls.ConnectionState {
	return nil
}
func (t *tcpTransport) RemoteAddress() string {
	return t.ch.RemoteAddr().String()
}
func (t *tcpTransport) Channel() listener.TimeoutReadWriteCloser {
	return t.ch
}
func (t *tcpTransport) Close() error {
	return t.ch.Close()
}

func NewTCPTransport(port int, ch chan<- listener.Transport) (io.Closer, error) {
	listener := &tcp{
		port: port,
	}
	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", listener.port))
	if err != nil {
		return nil, err
	}
	proxyListener := &proxyproto.Listener{Listener: tcp}
	listener.listener = proxyListener
	go listener.acceptLoop(ch)
	return proxyListener, nil
}

func (t *tcp) acceptLoop(ch chan<- listener.Transport) {
	var tempDelay time.Duration
	for {
		c, err := t.listener.Accept()
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
				log.Printf("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			log.Printf("connection handling failed: %v", err)
			t.listener.Close()
			return
		}
		t.queueSession(c, ch)
	}
}

func (t *tcp) queueSession(c net.Conn, ch chan<- listener.Transport) {
	ch <- &tcpTransport{
		ch: c,
	}
}
