package endpoints

import (
	"context"
	"time"

	"github.com/vx-labs/mqtt-broker/adapters/discovery/pb"
)

func (a *api) GetEndpoints(ctx context.Context, input *pb.GetEndpointsInput) (*pb.GetEndpointsOutput, error) {
	services, err := a.mesh.EndpointsByService(input.ServiceName)
	if err != nil {
		return nil, err
	}
	return &pb.GetEndpointsOutput{
		NodeServices: services,
	}, nil
}
func (a *api) AddServiceTag(ctx context.Context, input *pb.AddServiceTagInput) (*pb.AddServiceTagOutput, error) {
	err := a.mesh.AddServiceTag(input.ServiceID, input.TagKey, input.TagValue)
	if err != nil {
		return nil, err
	}
	return &pb.AddServiceTagOutput{}, nil
}
func (a *api) RemoveServiceTag(ctx context.Context, input *pb.RemoveServiceTagInput) (*pb.RemoveServiceTagOutput, error) {
	err := a.mesh.RemoveServiceTag(input.ServiceID, input.TagKey)
	if err != nil {
		return nil, err
	}
	return &pb.RemoveServiceTagOutput{}, nil
}

func (a *api) StreamEndpoints(input *pb.GetEndpointsInput, stream pb.DiscoveryService_StreamEndpointsServer) error {
	// TODO: use events
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		services, err := a.mesh.EndpointsByService(input.ServiceName)
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
