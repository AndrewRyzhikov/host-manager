package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"hostManager/internal/config"
	"hostManager/internal/service"
	api "hostManager/pkg/gen"
)

type Handler struct {
	api.UnimplementedDNSHostnameServiceServer
	manager service.HostManager
}

func NewHandler(manager service.HostManager) *Handler {
	return &Handler{manager: manager}
}

func Register(gRPC *grpc.Server, cfg config.BackupConfig) {
	manager := service.NewFileSystemHostManager(cfg)
	server := NewHandler(manager)

	api.RegisterDNSHostnameServiceServer(gRPC, server)
}

func (s *Handler) SetHostname(ctx context.Context, r *api.SetHostnameRequest) (*api.SetHostnameResponse, error) {
	if r.GetHostname() == "" {
		return nil, status.Error(codes.InvalidArgument, "hostname or hostname is empty")
	}

	if err := s.manager.SetHostname(ctx, r.GetHostname()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SetHostnameResponse{}, nil
}

func (s *Handler) ListDNSServers(ctx context.Context, r *api.ListDNSServersRequest) (*api.ListDNSServersResponse, error) {
	servers, err := s.manager.ListDNSServers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.ListDNSServersResponse{DnsServers: servers}, nil
}

func (s *Handler) AddDNSServer(ctx context.Context, r *api.AddDNSServerRequest) (*api.AddDNSServerResponse, error) {
	if r.GetDnsServer() == "" {
		return nil, status.Error(codes.InvalidArgument, "dns server is empty")
	}

	if err := s.manager.AddDNSServer(ctx, r.GetDnsServer()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.AddDNSServerResponse{}, nil
}

func (s *Handler) RemoveDNSServer(ctx context.Context, r *api.RemoveDNSServerRequest) (*api.RemoveDNSServerResponse, error) {
	if r.GetDnsServer() == "" {
		return nil, status.Error(codes.InvalidArgument, "dns server is empty")
	}

	if err := s.manager.RemoveDNSServer(ctx, r.GetDnsServer()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.RemoveDNSServerResponse{}, nil
}
