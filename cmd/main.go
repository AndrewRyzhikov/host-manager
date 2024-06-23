package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"

	cfg "hostManager/internal/config"
	"hostManager/internal/transport/grpc"
	"hostManager/internal/transport/rest"
)

func setupLog(config cfg.LogConfig) (zerolog.Logger, error) {
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

const DefaultConfigPath = "./config/local.yaml"

var config *cfg.Config
var ConfigPath = DefaultConfigPath

var rootCmd = &cobra.Command{
	Use: "host-manager",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		config, err = cfg.Load(ConfigPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config file")
		}

		log.Logger, err = setupLog(config.LogConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to setup logger")
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run host-manager",
	Run: func(cmd *cobra.Command, args []string) {
		grpcServer := grpc.NewServer(config.GRPCConfig.Port, log.Logger)
		if err := grpcServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start gRPC server")
		}
		httpServer := rest.NewServer(config.HTTPConfig, log.Logger)
		if err := httpServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
		log.Info().Msg("App started")

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

		<-stop

		if err := httpServer.Stop(); err != nil {
			log.Fatal().Err(err).Msg("Failed to stop HTTP server")
		}

		log.Info().Msg("http server stopped")

		grpcServer.Stop()
		log.Info().Msg("grpc server stopped")

		log.Info().Msg("App stopped")
	},
}

func main() {
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", DefaultConfigPath, fmt.Sprintf("config file path (default %s)", DefaultConfigPath))

	rootCmd.Execute()
}
