package pb

import "errors"

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/services/subscriptions/pb/ --go_out=plugins=grpc:. types.proto
var (
	ErrSubscriptionNotFound = errors.New("subscription not found in store")
)
