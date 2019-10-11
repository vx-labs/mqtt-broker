package network

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

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

/*
func unaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	fmt.Printf("RPC: %s, elapsed: %s, err: %v\n", method, time.Since(start), err)
	return err
}
*/

func GRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		//grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		//grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		/*		grpc.KeepaliveEnforcementPolicy(kaep),
				grpc.KeepaliveParams(kasp),*/
	}
}
func GRPCClientOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithInsecure(),
		//	grpc.WithKeepaliveParams(kacp),
		//	grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
		//	grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithBalancerName(roundrobin.Name),
	}
}
