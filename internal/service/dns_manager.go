package service

import "context"

type DNSManager interface {
	SetHostname(ctx context.Context, hostname string) error
	ListDNSServers(ctx context.Context) ([]string, error)
	AddDNSServer(ctx context.Context, server string) error
	RemoveDNSServer(ctx context.Context, server string) error
}
