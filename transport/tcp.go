package transport

import (
	"fmt"
	"log"
	"net"
	"time"

	proxyproto "github.com/armon/go-proxyproto"
)

type tcp struct {
	listener net.Listener
}

func NewTCPTransport(port int, handler func(Metadata) error) (net.Listener, error) {
	listener := &tcp{}
	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	proxyListener := &proxyproto.Listener{Listener: tcp}
	listener.listener = proxyListener
	go listener.acceptLoop(handler)
	return proxyListener, nil
}

func (t *tcp) acceptLoop(handler func(Metadata) error) {
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
		t.queueSession(c, handler)
	}
}

func (t *tcp) queueSession(c net.Conn, handler func(Metadata) error) {
	go handler(Metadata{
		Channel:         c,
		Encrypted:       false,
		EncryptionState: nil,
		Name:            "tcp",
		RemoteAddress:   c.RemoteAddr().String(),
	})
}
