package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type Env string

const (
	EnvDev     Env = "dev"
	EnvNonProd Env = "nonprod"
	EnvProd    Env = "prod"
)

type ServerConfig struct {
	Addr         string `mapstructure:"addr"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type WatchConfig struct {
	PollInterval       int    `mapstructure:"poll_interval"`
	FilterNames        string `mapstructure:"filter_names"`
	FilterPackageTypes string `mapstructure:"filter_package_types"`
	StateFile          string `mapstructure:"state_file"`
	MaxConcurrent      int    `mapstructure:"max_concurrent"`
	EnableTUI          bool   `mapstructure:"enable_tui"`
}

type Config struct {
	MCPRegistryURL string       `mapstructure:"mcp_registry_url"`
	OutputDir      string       `mapstructure:"output_dir"`
	LogLevel       LogLevel     `mapstructure:"log_level"`
	Env            Env          `mapstructure:"env"`
	Server         ServerConfig `mapstructure:"server"`
	Watch          WatchConfig  `mapstructure:"watch"`
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.nomad-mcp-pack/")

	viper.SetDefault("mcp_registry_url", "https://registry.modelcontextprotocol.io")
	viper.SetDefault("output_dir", "./packs")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("env", "prod")
	viper.SetDefault("server.addr", ":8080")
	viper.SetDefault("server.read_timeout", 10)
	viper.SetDefault("server.write_timeout", 10)
	viper.SetDefault("watch.poll_interval", 300)
	viper.SetDefault("watch.filter_names", "")
	viper.SetDefault("watch.filter_package_types", "npm,pypi,oci,nuget")
	viper.SetDefault("watch.state_file", "watch.json")
	viper.SetDefault("watch.max_concurrent", 5)
	viper.SetDefault("watch.enable_tui", false)

	viper.SetEnvPrefix("NOMAD_MCP_PACK")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
}

func GetConfig() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	switch cfg.LogLevel {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
	default:
		return nil, fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", cfg.LogLevel)
	}

	switch cfg.Env {
	case EnvDev, EnvNonProd, EnvProd:
	default:
		return nil, fmt.Errorf("invalid env: %s (must be dev, nonprod, or prod)", cfg.Env)
	}

	return &cfg, nil
}

func (c *Config) IsDev() bool {
	return c.Env == EnvDev
}

func (c *Config) IsNonProd() bool {
	return c.Env == EnvNonProd
}

func (c *Config) IsProd() bool {
	return c.Env == EnvProd
}
