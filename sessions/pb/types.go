package pb

import "errors"

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/sessions/pb/ --go_out=plugins=grpc:. types.proto
var (
	ErrSessionNotFound = errors.New("session not found in store")
)
