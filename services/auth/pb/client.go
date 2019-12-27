package pb

import (
	context "context"

	"google.golang.org/grpc"
)

type Client struct {
	api AuthServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		api: NewAuthServiceClient(conn),
	}
}

func (c *Client) CreateToken(ctx context.Context, protocol ProtocolContext, transport TransportContext) (*CreateTokenOutput, error) {
	return c.api.CreateToken(ctx, &CreateTokenInput{
		Protocol:  &protocol,
		Transport: &transport,
	})
}
