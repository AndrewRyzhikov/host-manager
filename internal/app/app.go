package app

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"hostManager/internal/app/grpc"
	"hostManager/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func StartApp() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config file")
	}

	log.Logger, err = setupLog(&cfg.LogConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup logger")
	}

	app := grpc.NewApp(cfg.GRPCConfig.Port, log.Logger)
	if err = app.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start app")
	}
	log.Info().Msg("App started...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	app.Stop()
	log.Info().Msg("App stopped")
}

func setupLog(config *config.LogConfig) (zerolog.Logger, error) {
	lumberjackCfg := config.Lumberjack

	lr := &lumberjack.Logger{
		Filename:   config.Path,
		MaxSize:    int(lumberjackCfg.MaxSize),
		MaxAge:     int(lumberjackCfg.MaxAge),
		MaxBackups: int(lumberjackCfg.MaxBackups),
		LocalTime:  lumberjackCfg.LocalTime,
		Compress:   lumberjackCfg.Compress,
	}

	level, err := zerolog.ParseLevel(config.Level)

	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("invalid log level: %s", config.Level)
	}

	return zerolog.New(lr).Level(level).With().Timestamp().Logger(), nil
}
