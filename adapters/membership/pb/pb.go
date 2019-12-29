package pb

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/adapters/membership/pb/ --go_out=plugins=grpc:. pb.proto

// APState represents a CRDT state store, that will be distributed over the mesh network.
type APState interface {
	Merge(inc []byte, full bool) error
	MarshalBinary() []byte
}

// Channel allows clients to send messages for a specific state type that will be
// broadcasted in a best-effort manner.
type Channel interface {
	Broadcast(b []byte)
	BroadcastFullState(b []byte)
}
