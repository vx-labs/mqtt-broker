package cluster

import (
	"github.com/hashicorp/memberlist"
	"google.golang.org/grpc"
)

type mockedChannel struct{}

func (m *mockedChannel) Broadcast([]byte) {}

type mockedMesh struct{}

func (m *mockedMesh) AddState(key string, state State) (Channel, error) {
	return &mockedChannel{}, nil
}
func (m *mockedMesh) Join(hosts []string) {}
func (m *mockedMesh) MemberRPCAddress(string) (string, error) {
	return "", ErrNodeNotFound
}
func (m *mockedMesh) ID() string {
	return "id"
}
func (m *mockedMesh) Peers() PeerStore {
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
func (m *mockedMesh) Leave()                  {}
func (m *mockedMesh) DiscoverPeers(PeerStore) {}
func (m *mockedMesh) Members() []*memberlist.Node {
	return nil
}

func MockedMesh() *mockedMesh {
	return &mockedMesh{}
}
