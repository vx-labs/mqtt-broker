package pb

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/adapters/discovery/pb/ --go_out=plugins=grpc:. pb.proto

type MembershipAdapter interface {
	Shutdown() error
	UpdateMetadata([]byte)
	OnNodeJoin(f func(id string, meta []byte))
	OnNodeUpdate(f func(id string, meta []byte))
	OnNodeLeave(f func(id string, meta []byte))
}
