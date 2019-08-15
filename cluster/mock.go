package cluster

import (
	"context"
	"errors"

	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/peers"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"google.golang.org/grpc"
)

var (
	ErrStateKeyAlreadySet = errors.New("specified key is already taken")
	ErrNodeNotFound       = errors.New("specified node not found in mesh")
)

type mockedChannel struct{}

func (m *mockedChannel) Broadcast([]byte) {}

type mockedMesh struct{}

func (m *mockedMesh) AddState(key string, state types.GossipState) (types.Channel, error) {
	return &mockedChannel{}, nil
}
func (m *mockedMesh) OnNodeLeave(f func(id string, meta pb.NodeMeta)) {}
func (m *mockedMesh) Join(hosts []string) error {
	return nil
}
func (m *mockedMesh) MemberRPCAddress(string) (string, error) {
	return "", ErrNodeNotFound
}
func (m *mockedMesh) ID() string {
	return "id"
}
func (m *mockedMesh) Status(context.Context, *pb.StatusInput) (*pb.StatusOutput, error) {
	return nil, nil
}
func (m *mockedMesh) Peers() peers.PeerStore {
	return nil
}

func (m *mockedMesh) DialAddress(service, id string, f func(*grpc.ClientConn) error) error {
	return nil
}

func (m *mockedMesh) DialService(id string) (*grpc.ClientConn, error) {
	return nil, nil
}
func (m *mockedMesh) RegisterService(name, addr string) error {
	return nil
}
func (m *mockedMesh) Health() string {
	return "ok"
}
func (m *mockedMesh) Leave()                        {}
func (m *mockedMesh) DiscoverPeers(peers.PeerStore) {}
func (m *mockedMesh) Members() []*memberlist.Node {
	return nil
}

func MockedMesh() *mockedMesh {
	return &mockedMesh{}
}
