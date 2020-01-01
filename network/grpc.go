package network

import (
	"crypto/tls"
	"os"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	"google.golang.org/grpc/credentials"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

func init() {
	grpc_prometheus.EnableHandlingTimeHistogram()
}

// From: https://github.com/grpc/grpc-go/blob/master/examples/features/keepalive/server/main.go
var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	Time:    5 * time.Second, // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout: 1 * time.Second, // Wait 1 second for the ping ack before assuming the connection is dead
}

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             5 * time.Second,  // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

func GRPCServerOptions() []grpc.ServerOption {
	tlsCreds, err := credentials.NewServerTLSFromFile(os.Getenv("TLS_CERTIFICATE"), os.Getenv("TLS_PRIVATE_KEY"))
	if err != nil {
		panic(err)
	}
	return []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.Creds(tlsCreds),
		/*		grpc.KeepaliveEnforcementPolicy(kaep),
				grpc.KeepaliveParams(kasp),*/
	}
}
func GRPCClientOptions() []grpc.DialOption {
	var tlsConfig credentials.TransportCredentials
	var err error
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(200*time.Millisecond, 0.3)),
		grpc_retry.WithMax(5),
	}
	if caCertificate := os.Getenv("TLS_CA_CERTIFICATE"); caCertificate != "" {
		tlsConfig, err = credentials.NewClientTLSFromFile(os.Getenv("TLS_CA_CERTIFICATE"), "")
		if err != nil {
			panic(err)
		}
	} else {
		tlsConfig = credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: os.Getenv("TLS_INSECURE_SKIP_VERIFY") == "true",
		})
	}
	return []grpc.DialOption{
		grpc.WithTransportCredentials(tlsConfig),
		//	grpc.WithKeepaliveParams(kacp),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 300 * time.Millisecond}),
		grpc.WithStreamInterceptor(
			grpc_middleware.ChainStreamClient(
				grpc_prometheus.StreamClientInterceptor,
				grpc_retry.StreamClientInterceptor(opts...),
			),
		),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				grpc_prometheus.UnaryClientInterceptor,
				grpc_retry.UnaryClientInterceptor(opts...),
			),
		),
		grpc.WithBalancerName(roundrobin.Name),
	}
}
