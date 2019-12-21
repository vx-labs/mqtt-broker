package events

import (
	"context"
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/messages/pb"
)

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/events/ --go_out=plugins=grpc:. events.proto

var (
	ErrNodeNotReady = errors.New("node is not ready")
)

func Decode(payload []byte) ([]*StateTransition, error) {
	format := StateTransitionSet{}
	err := proto.Unmarshal(payload, &format)
	if err != nil {
		return nil, err
	}
	return format.Events, nil
}
func Encode(events ...*StateTransition) ([]byte, error) {
	format := StateTransitionSet{
		Events: events,
	}
	return proto.Marshal(&format)
}

func Commit(ctx context.Context, client *pb.Client, shardKey string, events ...*StateTransition) error {
	if client == nil {
		return ErrNodeNotReady
	}
	payload, err := Encode(events...)
	if err != nil {
		return err
	}
	return client.Put(ctx, "events", shardKey, payload)
}
