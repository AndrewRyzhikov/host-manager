package grpc

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	grpc2 "hostManager/internal/transport/grpc"
	"net"
)

type App struct {
	post       int
	log        zerolog.Logger
	grpcServer *grpc.Server
}

func NewApp(post int, log zerolog.Logger) *App {
	grpcServer := grpc.NewServer()

	grpc2.Register(grpcServer)

	return &App{post: post, log: log, grpcServer: grpcServer}
}

func (a *App) Start() error {
	const op = "app.Start"

	a.log = a.log.With().Str("op", op).Logger()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.post))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info().Msg("grpc server is running on " + l.Addr().String())

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "app.Stop"

	a.log = a.log.With().Str("op", op).Logger()

	a.grpcServer.GracefulStop()
}
