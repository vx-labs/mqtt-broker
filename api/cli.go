package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/vx-labs/mqtt-broker/cluster/types"
)

func prefixVersion(suffix string) string {
	return fmt.Sprintf("/v1/%s", suffix)
}
func httpFail(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(`{"status": "InternalServerError", "status_code": 500, "error": "%s"}`, err.Error())))
}
func (b *api) acceptLoop(listener net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc(prefixVersion("sessions/"), func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		sessions, err := b.brokerClient.ListSessions(r.Context())
		if err != nil {
			httpFail(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(sessions)
	})
	mux.HandleFunc(prefixVersion("peers/"), func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		peers, err := b.mesh.Peers().All()
		if err != nil {
			httpFail(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(peers)
	})
	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		mux.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.String(), r.RemoteAddr)
	}))
}

func (b *api) Serve(_ int) net.Listener {
	if b.config.TlsPort > 0 {
		if b.config.TlsConfig != nil {
			ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", b.config.TlsPort), b.config.TlsConfig)
			if err != nil {
				log.Fatalf("failed to start TLS listener: %v", err)
			}
			b.listeners = append(b.listeners, ln)
			log.Printf("INFO: started TLS listener on port %d", b.config.TlsPort)
		} else {
			log.Printf("WARN: not starting TLS listener: TLS config is empty")
		}
	}
	if b.config.TcpPort > 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", b.config.TcpPort))
		if err != nil {
			log.Fatalf("failed to start TCP listener: %v", err)
		}
		b.listeners = append(b.listeners, ln)
		log.Printf("INFO: started TCP listener on port %d", b.config.TcpPort)
	}
	for _, lis := range b.listeners {
		b.acceptLoop(lis)
	}
	return nil
}
func (b *api) Shutdown() {
	for _, lis := range b.listeners {
		lis.Close()
	}
}
func (b *api) JoinServiceLayer(layer types.ServiceLayer) {
}
func (m *api) Health() string {
	return "ok"
}
