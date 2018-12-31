package transport

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/vx-labs/mqtt-broker/broker/listener"
)

type Conn struct {
	conn      net.Conn
	reader    *wsutil.Reader
	writer    *wsutil.Writer
	state     tls.ConnectionState
	opHandler wsutil.FrameHandler
}

func (c *Conn) Read(b []byte) (int, error) {
	n, err := c.reader.Read(b)
	if err == io.EOF || err == wsutil.ErrNoFrameAdvance {
		for {
			header, err := c.reader.NextFrame()
			if err != nil {
				return n, err
			}
			if header.OpCode.IsData() {
				return c.Read(b)
			}
			if header.OpCode.IsControl() {
				if err = c.opHandler(header, c.reader); err != nil {
					log.Printf("websocket control op handler failed: %v", err)
					return n, io.EOF
				}
			}
		}
	}

	return n, err
}
func (c *Conn) Write(b []byte) (int, error) {
	n, err := c.writer.Write(b)
	if err == nil {
		return n, c.writer.Flush()
	}
	return n, err
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
	port     int
	listener net.Listener
}

type wsTransport struct {
	ch net.Conn
}

func (t *wsTransport) Name() string {
	return "ws"
}

func (t *wsTransport) Encrypted() bool {
	return false
}
func (t *wsTransport) EncryptionState() *tls.ConnectionState {
	return nil
}
func (t *wsTransport) RemoteAddress() string {
	return t.ch.RemoteAddr().String()
}
func (t *wsTransport) Channel() listener.TimeoutReadWriteCloser {
	return t.ch
}
func (t *wsTransport) Close() error {
	return t.ch.Close()
}

func NewWSTransport(port int, ch chan<- listener.Transport) (io.Closer, error) {
	listener := &wsListener{
		port: port,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mqtt", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("INFO: starting websocket negociation with %s", r.RemoteAddr)
		conn, _, _, err := ws.UpgradeHTTP(r, w, http.Header{
			"Sec-WebSocket-Protocol": {"mqtt"},
		})
		if err != nil {
			log.Printf("ERR: websocket negociation with %s failed: %v", r.RemoteAddr, err)
			return
		}
		var (
			state  = ws.StateServerSide
			reader = wsutil.NewReader(conn, state)
			writer = wsutil.NewWriter(conn, state, ws.OpBinary)
		)
		listener.queueSession(&Conn{
			conn:      conn,
			reader:    reader,
			writer:    writer,
			opHandler: wsutil.ControlHandler(conn, state),
		}, ch)
	})
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to start WS listener: %v", err)
	}
	listener.listener = ln
	go http.Serve(ln, mux)
	return ln, nil
}

func (t *wsListener) queueSession(c *Conn, ch chan<- listener.Transport) {
	ch <- &wsTransport{
		ch: c,
	}
}
