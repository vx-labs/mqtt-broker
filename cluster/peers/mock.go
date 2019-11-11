package peers

import (
	"context"
	"errors"

	"github.com/hashicorp/memberlist"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
	"github.com/vx-labs/mqtt-broker/cluster/types"
	"google.golang.org/grpc"
)

type mockedChannel struct{}

func (m *mockedChannel) Broadcast([]byte)          {}
func (m *mockedChannel) BroadcastFullState([]byte) {}

type mockedMesh struct{}

func (m *mockedMesh) AddState(key string, state types.GossipState) (types.Channel, error) {
	return &mockedChannel{}, nil
}
func (m *mockedMesh) OnNodeLeave(f func(id string, meta pb.NodeMeta)) {}
func (m *mockedMesh) Join(hosts []string) error {
	return nil
}
func (m *mockedMesh) SendEvent(ctx context.Context, input *pb.SendEventInput) (*pb.SendEventOutput, error) {
	return &pb.SendEventOutput{}, nil
}
func (m *mockedMesh) MemberRPCAddress(string) (string, error) {
	return "", errors.New("node found found")
}
func (m *mockedMesh) ID() string {
	return "id"
}
func (m *mockedMesh) Health() string {
	return "ok"
}
func (m *mockedMesh) Status(context.Context, *pb.StatusInput) (*pb.StatusOutput, error) {
	return nil, nil
}
func (m *mockedMesh) Peers() PeerStore {
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
func (m *mockedMesh) PrepareShutdown(context.Context, *pb.PrepareShutdownInput) (*pb.PrepareShutdownOutput, error) {
	return nil, nil
}
func (m *mockedMesh) Members() []*memberlist.Node {
	return nil
}

func MockedMesh() *mockedMesh {
	return &mockedMesh{}
}
