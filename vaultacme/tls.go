package vaultacme

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

const globalPrefix = "mqtt/acme"

type vaultCache struct {
	api *api.Client
}

func (c *vaultCache) Get(ctx context.Context, key string) ([]byte, error) {
	path := fmt.Sprintf("secret/data/%s/%s", globalPrefix, key)
	response, err := c.api.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, autocert.ErrCacheMiss
	}
	data := response.Data["data"]
	if data == nil {
		return nil, autocert.ErrCacheMiss
	}
	kv := data.(map[string]interface{})
	if (kv["raw"]) == nil {
		return nil, autocert.ErrCacheMiss
	}
	return base64.StdEncoding.DecodeString(kv["raw"].(string))
}
func (c *vaultCache) Put(ctx context.Context, key string, data []byte) error {
	path := fmt.Sprintf("secret/data/%s/%s", globalPrefix, key)
	log.Printf("PUT %s", path)
	_, err := c.api.Logical().Write(path, map[string]interface{}{
		"data": map[string]interface{}{
			"raw": data,
		},
	})
	return err
}
func (c *vaultCache) Delete(ctx context.Context, key string) error {
	path := fmt.Sprintf("secret/data/%s/%s", globalPrefix, key)
	c.api.Logical().Delete(path)
	return nil
}

func loadVaultToken(api *api.Client) {
	fallback := func() {
		api.SetToken(os.Getenv("VAULT_TOKEN"))
	}
	_, err := os.Stat("secrets/vault_token")
	if err != nil {
		fallback()
		return
	}
	token, err := ioutil.ReadFile("secrets/vault_token")
	if err != nil {
		fallback()
		return
	}
	api.SetToken(string(token))
	sigUsr1 := make(chan os.Signal, 1)
	signal.Notify(sigUsr1, syscall.SIGUSR1)
	go func() {
		<-sigUsr1
		log.Println("INFO: received SIGUSR1, reloading vault token")
		loadVaultToken(api)
	}()
}

func GetConfig(ctx context.Context, cn string, logger *zap.Logger) (*tls.Config, error) {
	config := api.DefaultConfig()
	config.HttpClient = http.DefaultClient
	vaultAPI, err := api.NewClient(config)
	if err != nil {
		logger.Error("failed to load vault client", zap.Error(err), zap.String("vault_address", config.Address))
		return nil, err
	}
	health, err := vaultAPI.Sys().Health()
	if err != nil {
		logger.Error("failed to connect to vault", zap.Error(err), zap.String("vault_address", config.Address))
		return nil, err
	}
	if health.Sealed || !health.Initialized {
		if err != nil {
			logger.Error("vault is not available", zap.Error(err), zap.Bool("vault_sealed", health.Sealed), zap.Bool("vault_initialized", health.Initialized), zap.String("vault_address", config.Address))
			return nil, err
		}
	}
	loadVaultToken(vaultAPI)
	_, err = vaultAPI.Sys().Leader()
	if err != nil {
		logger.Error("failed to dialog with to vault", zap.Error(err), zap.String("vault_address", config.Address))
		return nil, err
	}

	logger.Debug("using vault as ACME storage", zap.String("vault_address", config.Address))

	manager := &autocert.Manager{
		Email:  os.Getenv("LE_EMAIL"),
		Prompt: autocert.AcceptTOS,
		HostPolicy: func(ctx context.Context, host string) error {
			if host != cn {
				return fmt.Errorf("host %s not allowed", host)
			}
			return nil
		},
		Cache: &vaultCache{
			api: vaultAPI,
		},
		Client: &acme.Client{},
	}
	logger.Debug("loaded ACME client", zap.Strings("acme_common_names", []string{cn}))
	tlsConfig := manager.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS12
	return tlsConfig, nil
}
