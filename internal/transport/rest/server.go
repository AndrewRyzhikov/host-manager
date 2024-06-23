package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"hostManager/internal/config"
	"hostManager/pkg/gen"
)

type Server struct {
	httpServer *http.Server
	cfg        config.HTTPConfig
	log        zerolog.Logger
}

func NewServer(cfg config.HTTPConfig, log zerolog.Logger) *Server {
	return &Server{cfg: cfg, log: log}
}

func (s *Server) Start() error {
	const op = "http.Start"
	s.log = s.log.With().Str("op", op).Logger()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := gen.RegisterDNSHostnameServiceHandlerFromEndpoint(context.Background(), mux, s.cfg.GRPCServerEndpoint, opts)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.httpServer = &http.Server{
		Addr:    ":" + s.cfg.Port,
		Handler: mux,
	}

	s.log.Info().Msgf("http server started on %s", s.httpServer.Addr)

	go func() {
		if err = s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("http server start failed")
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	const op = "http.Stop"

	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
