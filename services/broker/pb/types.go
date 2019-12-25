package pb

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/services/broker/pb --go_out=plugins=grpc:. pb.proto
