package crdt

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/crdt/ --go_out=plugins=grpc:. types.proto

var _ Entry = &MockedEntry{}
