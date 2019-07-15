package transport

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Conn struct {
	conn   net.Conn
	reader io.Reader
	state  tls.ConnectionState
}

func (c *Conn) nextFrame() error {
	buf, err := wsutil.ReadClientBinary(c.conn)
	if err != nil {
		return err
	}
	c.reader = bytes.NewBuffer(buf)
	return nil
}
func (c *Conn) Read(b []byte) (int, error) {
	n := 0
	var err error
	for {
		if c.reader == nil {
			err := c.nextFrame()
			if err != nil {
				return 0, err
			}
		}
		n, err = c.reader.Read(b)
		if err == io.EOF {
			c.reader = nil
			if len(b) > n {
				continue
			}
			return n, nil
		}
		return n, err
	}
}
func (c *Conn) Write(b []byte) (int, error) {
	return len(b), wsutil.WriteServerBinary(c.conn, b)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

type wsListener struct {
	listener net.Listener
}

func NewWSTransport(port int, handler func(Metadata) error) (net.Listener, error) {
	listener := &wsListener{}

	mux := http.NewServeMux()
	mux.HandleFunc("/mqtt", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("INFO: starting websocket negociation with %s", r.RemoteAddr)
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Printf("ERR: websocket negociation with %s failed: %v", r.RemoteAddr, err)
			return
		}

		listener.queueSession(&Conn{
			conn: conn,
		}, handler)
	})
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to start WS listener: %v", err)
	}
	listener.listener = ln
	go http.Serve(ln, mux)
	return ln, nil
}

func (t *wsListener) queueSession(c *Conn, handler func(Metadata) error) {
	handler(Metadata{
		Channel:         c,
		Encrypted:       false,
		EncryptionState: nil,
		Name:            "ws",
		RemoteAddress:   c.RemoteAddr().String(),
	})
}
