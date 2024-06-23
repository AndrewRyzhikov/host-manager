package grpc

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"hostManager/internal/service"
	api "hostManager/pkg/gen"
)

type Server struct {
	api.UnimplementedDNSHostnameServiceServer
	manager service.DNSManager
}

func NewServer(manager service.DNSManager) *Server {
	return &Server{manager: manager}
}

func Register(gRPC *grpc.Server) {
	manager := service.NewFileSystemDNSManager(log.Logger)
	server := NewServer(manager)

	api.RegisterDNSHostnameServiceServer(gRPC, server)
}

func (s *Server) SetHostname(ctx context.Context, r *api.SetHostnameRequest) (*api.Response, error) {
	if r.GetHostname() == "" {
		return nil, status.Error(codes.InvalidArgument, "hostname or hostname is empty")
	}

	if err := s.manager.SetHostname(ctx, r.GetHostname()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.Response{}, nil
}
func (s *Server) ListDNSServers(ctx context.Context, r *api.Request) (*api.ListDNSServersResponse, error) {
	servers, err := s.manager.ListDNSServers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.ListDNSServersResponse{DnsServers: servers}, nil
}

func (s *Server) AddDNSServer(ctx context.Context, r *api.AddDNSServerRequest) (*api.Response, error) {
	if r.GetDnsServer() == "" {
		return nil, status.Error(codes.InvalidArgument, "dns server is empty")
	}

	if err := s.manager.AddDNSServer(ctx, r.GetDnsServer()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.Response{}, nil
}
func (s *Server) RemoveDNSServer(ctx context.Context, r *api.RemoveDNSServerRequest) (*api.Response, error) {
	if r.GetDnsServer() == "" {
		return nil, status.Error(codes.InvalidArgument, "dns server is empty")
	}

	if err := s.manager.RemoveDNSServer(ctx, r.GetDnsServer()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.Response{}, nil
}
