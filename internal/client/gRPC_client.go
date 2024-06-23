package client

import "context"

type gRPCClient struct {
}

func newGRPCClient() *gRPCClient {

	return &gRPCClient{}
}

func (g gRPCClient) SetHostname(ctx context.Context, hostname string) error {
	//TODO implement me
	panic("implement me")
}

func (g gRPCClient) ListDNSServers(ctx context.Context) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (g gRPCClient) AddDNSServer(ctx context.Context, server string) error {
	//TODO implement me
	panic("implement me")
}

func (g gRPCClient) RemoveDNSServer(ctx context.Context, server string) error {
	//TODO implement me
	panic("implement me")
}
