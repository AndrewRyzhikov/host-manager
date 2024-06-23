package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	GRPCConfig   GRPCConfig   `yaml:"grpc" env-required:"true"`
	HTTPConfig   HTTPConfig   `yaml:"http" env-required:"true"`
	BackupConfig BackupConfig `yaml:"backup"`
	LogConfig    LogConfig    `yaml:"log" env-required:"true"`
}

type HTTPConfig struct {
	Port               string `yaml:"port" env-required:"true"`
	GRPCServerEndpoint string `yaml:"grpc_server_endpoint" env-required:"true"`
}

type GRPCConfig struct {
	Port int `yaml:"port" env-required:"true"`
}

type BackupConfig struct {
	BackupHostnameFilePath string `yaml:"backup_hostname_file_path" env-default:"/etc/backup/hostname/"`
	BackupDNSFilePath      string `yaml:"backup_dns_file_path" env-default:"/etc/backup/dns/"`
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

func Load() (*Config, error) {
	path := fetchConfig()
	if path == "" {
		return nil, errors.New("config file not exist")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config path does not exist: %s", path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}

func fetchConfig() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config")
	flag.Parse()

	return res
}
