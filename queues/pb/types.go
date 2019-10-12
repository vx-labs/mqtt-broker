package pb

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/queues/pb/ --go_out=plugins=grpc:. types.proto
