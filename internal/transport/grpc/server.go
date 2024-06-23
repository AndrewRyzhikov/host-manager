package grpc

import (
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"hostManager/internal/config"
)

type Server struct {
	post       int
	grpcServer *grpc.Server
	log        zerolog.Logger
}

func NewServer(post int, log zerolog.Logger, cfg config.BackupConfig) *Server {
	grpcServer := grpc.NewServer()
	Register(grpcServer, cfg)

	return &Server{post: post, log: log, grpcServer: grpcServer}
}

func (s *Server) Start() error {
	const op = "grpc.Start"
	s.log = log.With().Str("op", op).Logger()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.post))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info().Msgf("gRPC server started on %s", l.Addr().String())

	go func() {
		if err := s.grpcServer.Serve(l); err != nil {
			s.log.Fatal().Msg("grpc server failed to serve")
		}
	}()

	return nil
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
