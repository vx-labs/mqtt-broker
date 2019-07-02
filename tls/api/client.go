package api

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"

	"github.com/go-acme/lego/challenge"

	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/providers/dns/cloudflare"
	"github.com/go-acme/lego/registration"
	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	config "github.com/vx-labs/iot-mqtt-config"
	"github.com/vx-labs/mqtt-broker/tls/cache"
	"golang.org/x/net/context"
)

type Client struct {
	api     *lego.Client
	cache   *cache.VaultProvider
	account *Account
}

type Account struct {
	key          *rsa.PrivateKey
	email        string
	Registration *registration.Resource
}

func (u Account) GetEmail() string {
	return u.email
}
func (u Account) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u Account) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func newHttpClient(httpConfig config.HTTPSchema) *http.Client {
	client := http.Client{}
	if httpConfig.Proxy != "" {
		proxyURL, err := url.Parse(httpConfig.Proxy)
		if err == nil {
			log.Printf("http-client: using proxy at %s", proxyURL.String())
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   15 * time.Second,
				ResponseHeaderTimeout: 15 * time.Second,
				ExpectContinueTimeout: 3 * time.Second,
			}
		}
	}
	return &client
}

func New(consulAPI *consul.Client, vaultAPI *vault.Client, o ...Opt) (*Client, error) {
	opts := getOpts(o)
	if opts.Email == "" {
		return nil, fmt.Errorf("missing email address")
	}
	ctx := context.Background()
	prefix := "tls"
	if opts.UseStaging {
		prefix = "tls-staging"
	}
	store := cache.NewVaultProvider(consulAPI, vaultAPI, prefix)
	lock, err := store.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()
	key, err := store.GetKey(ctx, "account")
	if err != nil {
		logrus.Warnf("failed to fetch account key from cache, generating a new one: %v", err)
		key, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, err
		}
		logrus.Infof("saving account private key")
		err = store.SaveKey(ctx, "account", key)
		if err != nil {
			logrus.Errorf("failed to save private key: %v", err)
			return nil, err
		}
	} else {
		logrus.Infof("fetched account private key from cache")
	}
	account := Account{
		email: opts.Email,
		key:   key,
	}
	reg, err := store.GetRegistration(ctx)
	if err != nil {
		logrus.Warnf("failed to fetch ACME account from cache")
	} else {
		account.Registration = reg
	}
	c := &Client{
		account: &account,
	}
	httpConfig, _, err := config.HTTP(consulAPI)
	if err != nil {
		return nil, err
	}
	httpClient := newHttpClient(httpConfig)
	legoConfig := &lego.Config{
		HTTPClient: httpClient,
		User:       &account,
	}
	var client *lego.Client
	if opts.UseStaging {
		legoConfig.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	} else {
		legoConfig.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	}
	if err != nil {
		return nil, err
	}

	c.api, err = lego.NewClient(legoConfig)
	if err != nil {
		return nil, err
	}
	cfCreds, err := config.Cloudflare(vaultAPI)
	if err != nil {
		return nil, err
	}
	cfConfig := cloudflare.NewDefaultConfig()
	cfConfig.TTL = 125
	cfConfig.HTTPClient = newHttpClient(httpConfig)
	cfConfig.AuthEmail = cfCreds.EmailAddress
	cfConfig.AuthKey = cfCreds.APIToken
	cf, err := cloudflare.NewDNSProviderConfig(cfConfig)
	if err != nil {
		return nil, err
	}
	c.cache = store

	if account.Registration == nil {
		reg, err = client.Registration.Register(registration.RegisterOptions{
			TermsOfServiceAgreed: true,
		})
		if err != nil {
			return nil, err
		}
		err = store.SaveRegistration(ctx, reg)
		if err != nil {
			return nil, err
		}
	}
	c.api.Challenge.SetDNS01Provider(cf, dns01.AddDNSTimeout(5*time.Minute), dns01.WrapPreCheck(func(domain, fqdn, value string, check dns01.PreCheckFunc) (bool, error) {
		name := strings.TrimSuffix(fqdn, ".")
		log.Printf("Waiting for DNS propagation of record %s", name)
		result, err := net.LookupTXT(name)
		if err != nil {
			log.Printf("ERR: %v", err)
			return false, nil
		}
		if len(result) > 0 {
			return true, nil
		}
		return false, nil
	}))
	c.api.Challenge.Remove(challenge.HTTP01)
	c.api.Challenge.Remove(challenge.TLSALPN01)
	return c, err
}

func (c *Client) GetCertificate(ctx context.Context, cn string) ([]tls.Certificate, error) {
	l, err := c.cache.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer l.Unlock()
	key, err := c.cache.GetKey(ctx, cn)
	if err != nil {
		logrus.Infof("generating a new private key")
		key, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, err
		}
		err = c.cache.SaveKey(ctx, cn, key)
		if err != nil {
			return nil, err
		}
		logrus.Infof("generated and saved a new private key")
	} else {
		logrus.Infof("fetched key from cache")
	}
	cert, err := c.cache.GetCert(ctx, cn)
	if err != nil {
		logrus.Infof("request certificate from ACME")
		request := certificate.ObtainRequest{
			Domains:    []string{cn},
			Bundle:     true,
			PrivateKey: key,
		}
		certificates, err := c.api.Certificate.Obtain(request)
		if err != nil {
			return nil, err
		}
		cert = certificates.Certificate
		logrus.Infof("saving cert to cache")
		err = c.cache.SaveCert(ctx, cn, cert)
		if err != nil {
			logrus.Errorf("failed to save letsencrypt certs: %v", err)
			return nil, err
		}
	}
	encodedKey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	tlsCert, err := tls.X509KeyPair(cert, encodedKey)
	if err != nil {
		logrus.Errorf("could not load certificates: %v", err)
		return nil, err
	}
	return []tls.Certificate{
		tlsCert,
	}, nil
}
