package endpoints

import (
	"context"
	"time"

	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
)

func (a *api) GetEndpoints(ctx context.Context, input *pb.GetEndpointsInput) (*pb.GetEndpointsOutput, error) {
	services, err := a.mesh.EndpointsByService(input.ServiceName, "rpc")
	if err != nil {
		return nil, err
	}
	return &pb.GetEndpointsOutput{
		NodeServices: services,
	}, nil
}

func (a *api) StreamEndpoints(input *pb.GetEndpointsInput, stream pb.DiscoveryService_StreamEndpointsServer) error {
	// TODO: use events
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		services, err := a.mesh.EndpointsByService(input.ServiceName, "rpc")
		if err != nil {
			return err
		}
		err = stream.Send(&pb.GetEndpointsOutput{
			NodeServices: services,
		})
		if err != nil {
			return err
		}
		<-ticker.C
	}
}
