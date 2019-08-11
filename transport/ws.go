package transport

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/websocket"
)

type wsListener struct {
	listener net.Listener
}

func NewWSTransport(port int, handler func(Metadata) error) (net.Listener, error) {
	listener := &wsListener{}

	mux := http.NewServeMux()
	server := &websocket.Server{
		Handshake: func(config *websocket.Config, r *http.Request) error {
			log.Printf("new websocket connection from %s", r.RemoteAddr)
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
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to start WS listener: %v", err)
	}
	listener.listener = ln
	go http.Serve(ln, mux)
	return ln, nil
}

func (t *wsListener) queueSession(c *websocket.Conn, handler func(Metadata) error) {
	r := c.Request()
	handler(Metadata{
		Channel:         c,
		Encrypted:       false,
		EncryptionState: nil,
		Name:            "ws",
		RemoteAddress:   r.RemoteAddr,
	})
}
