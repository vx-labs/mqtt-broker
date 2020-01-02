package cobra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/network"

	"github.com/vx-labs/mqtt-broker/services/kv/pb"
	"google.golang.org/grpc"
)

type PeersAnswer []Peer
type Peer struct {
	ID             string
	HostedServices []struct {
		ID             string
		NetworkAddress string
		Tags           []struct {
			Key   string
			Value string
		}
	}
}

func resolveServiceAddress(discoveryURL string, name string) (string, error) {
	out := PeersAnswer{}
	resp, err := http.Get(fmt.Sprintf("%s/v1/peers", discoveryURL))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return "", err
	}
	for _, peer := range out {
		for _, service := range peer.HostedServices {
			if service.ID == name {
				return service.NetworkAddress, nil
			}
		}
	}
	return "", errors.New("service not found")
}

func getClient(config *viper.Viper) *pb.Client {
	host := config.GetString("host")
	if host == "" {
		endpoint := config.GetString("discovery-url")
		adapter := discovery.PB(context.TODO(), "", endpoint, zap.NewNop())
		conn, err := adapter.DialService("kv")
		if err != nil {
			log.Fatalf("failed to connect %s: %v", host, err)
		}
		return pb.NewClient(conn)
	}
	opts := network.GRPCClientOptions()
	conn, err := grpc.Dial(host,
		append(opts, grpc.WithTimeout(800*time.Millisecond))...)
	if err != nil {
		log.Fatalf("failed to connect %s: %v", host, err)
	}
	return pb.NewClient(conn)
}
