package state

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/state/ --go_out=plugins=grpc:. types.proto

var _ Entry = &MockedEntry{}
