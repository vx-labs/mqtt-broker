package peers

import (
	"errors"

	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"google.golang.org/grpc"
)

type mockedChannel struct{}

func (m *mockedChannel) Broadcast([]byte) {}

type mockedMesh struct{}

func (m *mockedMesh) AddState(key string, state types.State) (types.Channel, error) {
	return &mockedChannel{}, nil
}
func (m *mockedMesh) OnNodeLeave(f func(id string, meta pb.NodeMeta)) {}
func (m *mockedMesh) Join(hosts []string)                             {}
func (m *mockedMesh) MemberRPCAddress(string) (string, error) {
	return "", errors.New("node found found")
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
