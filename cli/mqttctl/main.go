package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	eventsCommand "github.com/vx-labs/mqtt-broker/events/cobra"
	authCommand "github.com/vx-labs/mqtt-broker/services/auth/cobra"
	kvCommand "github.com/vx-labs/mqtt-broker/services/kv/cobra"
	messagesCommand "github.com/vx-labs/mqtt-broker/services/messages/cobra"
	queuesCommand "github.com/vx-labs/mqtt-broker/services/queues/cobra"
	subscriptionsCommand "github.com/vx-labs/mqtt-broker/services/subscriptions/cobra"
)

func listLocalIP() []net.IP {
	out := []net.IP{}
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				out = append(out, v.IP)
			case *net.IPAddr:
				out = append(out, v.IP)
			}
		}
	}
	return out
}

func TLSHelper(config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "generate-tls",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("certificate-file", c.Flags().Lookup("certificate-file"))
			config.BindPFlag("private-key-file", c.Flags().Lookup("private-key-file"))
		},
		Run: func(cmd *cobra.Command, _ []string) {
			log.Printf("INFO: generating self-signed TLS certificate.")
			log.Printf("INFO: if this operation seems too long, check this host's entropy.")
			privkey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				log.Printf("ERR: §%v", err)
				return
			}
			certTemplate := &x509.Certificate{
				NotAfter:     time.Now().Add(12 * 30 * 24 * time.Hour),
				SerialNumber: big.NewInt(1),
				IPAddresses:  listLocalIP(),
				DNSNames:     []string{"*"},
				Subject: pkix.Name{
					CommonName: os.Getenv("HOSTNAME"),
				},
				ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
				KeyUsage:    x509.KeyUsageDigitalSignature,
			}

			certBody, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, privkey.Public(), privkey)
			if err != nil {
				log.Printf("ERR: §%v", err)
				return
			}
			certFile, err := os.Create(config.GetString("certificate-file"))
			if err != nil {
				log.Printf("ERR: %v", err)
				return
			}
			defer certFile.Close()
			keyFile, err := os.Create(config.GetString("private-key-file"))
			if err != nil {
				log.Printf("ERR: %v", err)
				return
			}
			defer keyFile.Close()
			pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privkey)})
			pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBody})
		},
	}
	c.Flags().StringP("certificate-file", "c", "./run_config/cert.pem", "Write certificate to this file")
	c.Flags().StringP("private-key-file", "k", "./run_config/privkey.pem", "Write private key to this file")
	return c
}

func main() {
	rootCmd := &cobra.Command{}
	config := viper.New()
	rootCmd.PersistentFlags().StringP("host", "", "", "remote GRPC endpoint")
	config.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	ctx := context.Background()
	messagesCommand.Register(ctx, rootCmd, config)
	kvCommand.Register(ctx, rootCmd, config)
	eventsCommand.Register(ctx, rootCmd, config)
	queuesCommand.Register(ctx, rootCmd, config)
	subscriptionsCommand.Register(ctx, rootCmd, config)
	authCommand.Register(ctx, rootCmd, config)
	rootCmd.AddCommand(TLSHelper(config))
	rootCmd.Execute()
}
