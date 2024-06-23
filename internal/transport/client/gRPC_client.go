package client

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	api "hostManager/pkg/gen"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client api.DNSHostnameServiceClient
}

func NewGRPCClient(serverAddr string) *GRPCClient {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to gRPC server")
	}

	client := api.NewDNSHostnameServiceClient(conn)

	return &GRPCClient{
		conn:   conn,
		client: client,
	}
}

func (g *GRPCClient) SetHostname(ctx context.Context, hostname string) error {
	if _, err := g.client.SetHostname(ctx, &api.SetHostnameRequest{Hostname: hostname}); err != nil {
		return err
	}

	return nil
}

func (g *GRPCClient) ListDNSServers(ctx context.Context) ([]string, error) {
	r, err := g.client.ListDNSServers(ctx, &api.ListDNSServersRequest{})
	if err != nil {
		return nil, err
	}

	return r.DnsServers, nil
}

func (g *GRPCClient) AddDNSServer(ctx context.Context, server string) error {
	if _, err := g.client.AddDNSServer(ctx, &api.AddDNSServerRequest{DnsServer: server}); err != nil {
		return err
	}

	return nil
}

func (g *GRPCClient) RemoveDNSServer(ctx context.Context, server string) error {
	if _, err := g.client.RemoveDNSServer(ctx, &api.RemoveDNSServerRequest{DnsServer: server}); err != nil {
		return err
	}

	return nil
}

func (g *GRPCClient) Close() error {
	if err := g.conn.Close(); err != nil {
		return err
	}

	return nil
}
