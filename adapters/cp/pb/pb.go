package pb

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/adapters/cp/pb/ --go_out=plugins=grpc:. pb.proto

import "io"

type CPState interface {
	Apply(event []byte) error
	Snapshot() io.ReadCloser
	Restore(io.Reader) error
}
