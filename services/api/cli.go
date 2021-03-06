package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/vaultacme"

	"go.uber.org/zap"
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
	mux.HandleFunc(prefixVersion("subscriptions/"), func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		sessions, err := b.subscriptionsClient.All(r.Context())
		if err != nil {
			httpFail(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(sessions)
	})
	mux.HandleFunc(prefixVersion("sessions/"), func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		sessions, err := b.sessionsClient.All(r.Context())
		if err != nil {
			httpFail(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(sessions)
	})
	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		mux.ServeHTTP(w, r)
		b.logger.Debug("served http request",
			zap.String("http_request_method", r.Method),
			zap.String("http_request_url", r.URL.String()),
			zap.String("remote_address", r.RemoteAddr),
			zap.Duration("request_duration", time.Since(now)))
	}))
}

func (b *api) Serve(_ int) net.Listener {
	return nil
}
func (b *api) Shutdown() {
	for _, lis := range b.listeners {
		lis.Close()
	}
}
func (b *api) Start(id, name string, catalog discovery.ServiceCatalog, logger *zap.Logger) error {
	if b.config.TlsPort > 0 {
		TLSConfig, err := vaultacme.GetConfig(b.ctx, b.config.TlsCommonName, b.logger)
		if err != nil {
			b.logger.Fatal("failed to load TLS config", zap.Error(err))
		}
		ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", b.config.TlsPort), TLSConfig)
		if err != nil {
			b.logger.Fatal("failed to start listener",
				zap.String("transport", "tls"), zap.Error(err))
		}
		b.listeners = append(b.listeners, ln)
		b.logger.Info("started listener",
			zap.String("transport", "tls"))
	}
	if b.config.TcpPort > 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", b.config.TcpPort))
		if err != nil {
			b.logger.Fatal("failed to start listener",
				zap.String("transport", "tcp"), zap.Error(err))
		}
		b.listeners = append(b.listeners, ln)
		b.logger.Info("started listener",
			zap.String("transport", "tcp"))
	}
	for _, lis := range b.listeners {
		b.acceptLoop(lis)
	}
	return nil
}
func (m *api) Health() (string, string) {
	return "ok", ""
}
