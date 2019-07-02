package cache

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/go-acme/lego/registration"
	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"golang.org/x/net/context"
)

const globalPrefix = "mqtt"

type VaultProvider struct {
	vault  *vault.Client
	consul *consul.Client
	prefix string
}

func NewVaultProvider(c *consul.Client, v *vault.Client, p string) *VaultProvider {
	return &VaultProvider{
		vault:  v,
		consul: c,
		prefix: fmt.Sprintf("%s/%s", globalPrefix, p),
	}
}

func (e *VaultProvider) savePrivateKey(ctx context.Context, path string, privkey *rsa.PrivateKey) error {
	encoded := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privkey),
	})
	_, err := e.vault.Logical().Write(path, map[string]interface{}{
		"data": map[string]interface{}{
			"private_key": encoded,
		},
	})
	return err
}

func (e *VaultProvider) getPrivateKey(ctx context.Context, path string) (*rsa.PrivateKey, error) {
	response, err := e.vault.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("key not found")
	}
	data := response.Data["data"]
	if data == nil {
		return nil, fmt.Errorf("key not found")
	}
	kv := data.(map[string]interface{})
	privkey, err := base64.StdEncoding.DecodeString(kv["private_key"].(string))
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(privkey))
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func (e *VaultProvider) SaveKey(ctx context.Context, cn string, privkey *rsa.PrivateKey) error {
	key := fmt.Sprintf("secret/data/%s/%s/private_key", e.prefix, cn)
	return e.savePrivateKey(ctx, key, privkey)
}
func (e *VaultProvider) GetKey(ctx context.Context, cn string) (*rsa.PrivateKey, error) {
	path := fmt.Sprintf("secret/data/%s/%s/private_key", e.prefix, cn)
	return e.getPrivateKey(ctx, path)
}

func (e *VaultProvider) SaveCert(ctx context.Context, cn string, cert []byte) error {
	key := fmt.Sprintf("secret/data/%s/%s/certificate", e.prefix, cn)
	_, err := e.vault.Logical().Write(key, map[string]interface{}{
		"data": map[string]interface{}{
			"certificate": cert,
		},
	})
	return err
}

func (e *VaultProvider) GetCert(ctx context.Context, cn string) ([]byte, error) {
	key := fmt.Sprintf("secret/data/%s/%s/certificate", e.prefix, cn)
	response, err := e.vault.Logical().Read(key)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("certificate not found")
	}
	data := response.Data["data"]
	if data == nil {
		return nil, fmt.Errorf("certificate not found")
	}
	kv := data.(map[string]interface{})
	return base64.StdEncoding.DecodeString(kv["certificate"].(string))
}
func (e *VaultProvider) Lock(ctx context.Context) (*consul.Lock, error) {
	l := NewConsulLocker(e.consul)
	return l.Lock(ctx, e.prefix)
}
func (e *VaultProvider) SaveRegistration(ctx context.Context, reg *registration.Resource) error {
	payload, err := json.Marshal(reg)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("secret/data/%s/account/registration", e.prefix)
	_, err = e.vault.Logical().Write(key, map[string]interface{}{
		"data": map[string]interface{}{
			"registration": payload,
		},
	})
	return err
}
func (e *VaultProvider) GetRegistration(ctx context.Context) (*registration.Resource, error) {
	key := fmt.Sprintf("secret/data/%s/account/registration", e.prefix)
	response, err := e.vault.Logical().Read(key)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("registration not found")
	}
	data := response.Data["data"]
	if data == nil {
		return nil, fmt.Errorf("registration not found")
	}
	kv := data.(map[string]interface{})
	payload, err := base64.StdEncoding.DecodeString(kv["registration"].(string))
	if err != nil {
		return nil, err
	}
	reg := &registration.Resource{}
	return reg, json.Unmarshal(payload, reg)
}
