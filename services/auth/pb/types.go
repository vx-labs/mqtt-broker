package pb

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/services/auth/pb --go_out=plugins=grpc:. pb.proto
