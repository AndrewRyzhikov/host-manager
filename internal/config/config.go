package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	GRPCConfig GRPCConfig `yaml:"grpc" env-required:"true"`
	HTTPConfig HTTPConfig `yaml:"http" env-required:"true"`
	LogConfig  LogConfig  `yaml:"log" env-required:"true"`
}

type HTTPConfig struct {
	Port               string `yaml:"port" env-required:"true"`
	GRPCServerEndpoint string `yaml:"grpc_server_endpoint" env-required:"true"`
}

type GRPCConfig struct {
	Port int `yaml:"port" env-required:"true"`
}

type LogConfig struct {
	Level      string           `yaml:"level" env-default:"INFO"`
	Path       string           `yaml:"path" env-required:"true"`
	Lumberjack LumberjackConfig `yaml:"lumberjack" env-required:"true"`
}

type LumberjackConfig struct {
	MaxSize    uint64 `yaml:"max_size"`
	MaxAge     uint64 `yaml:"max_age"`
	MaxBackups uint64 `yaml:"max_backups"`
	LocalTime  bool   `yaml:"local_time"`
	Compress   bool   `yaml:"compress"`
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config path does not exist: %s", path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}
