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
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type OutputType string

const (
	OutputTypePackdir OutputType = "packdir"
	OutputTypeArchive OutputType = "archive"
)

var DefaultConfig = struct {
	MCPRegistryURL          string
	OutputDir               string
	OutputType              string
	LogLevel                string
	Env                     string
	ServerAddr              string
	ServerReadTimeout       int
	ServerWriteTimeout      int
	WatchPollInterval       int
	WatchFilterNames        string
	WatchFilterPackageTypes string
	WatchStateFile          string
	WatchMaxConcurrent      int
	WatchEnableTUI          bool
}{
	MCPRegistryURL:          "https://registry.modelcontextprotocol.io",
	OutputDir:               "./packs",
	OutputType:              "packdir",
	LogLevel:                "info",
	Env:                     "prod",
	ServerAddr:              ":8080",
	ServerReadTimeout:       10,
	ServerWriteTimeout:      10,
	WatchPollInterval:       300,
	WatchFilterNames:        "",
	WatchFilterPackageTypes: "npm,pypi,oci,nuget",
	WatchStateFile:          "watch.json",
	WatchMaxConcurrent:      5,
	WatchEnableTUI:          false,
}

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
	OutputType     OutputType   `mapstructure:"output_type"`
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

	viper.SetDefault("mcp_registry_url", DefaultConfig.MCPRegistryURL)
	viper.SetDefault("output_dir", DefaultConfig.OutputDir)
	viper.SetDefault("output_type", DefaultConfig.OutputType)
	viper.SetDefault("log_level", DefaultConfig.LogLevel)
	viper.SetDefault("env", DefaultConfig.Env)
	viper.SetDefault("server.addr", DefaultConfig.ServerAddr)
	viper.SetDefault("server.read_timeout", DefaultConfig.ServerReadTimeout)
	viper.SetDefault("server.write_timeout", DefaultConfig.ServerWriteTimeout)
	viper.SetDefault("watch.poll_interval", DefaultConfig.WatchPollInterval)
	viper.SetDefault("watch.filter_names", DefaultConfig.WatchFilterNames)
	viper.SetDefault("watch.filter_package_types", DefaultConfig.WatchFilterPackageTypes)
	viper.SetDefault("watch.state_file", DefaultConfig.WatchStateFile)
	viper.SetDefault("watch.max_concurrent", DefaultConfig.WatchMaxConcurrent)
	viper.SetDefault("watch.enable_tui", DefaultConfig.WatchEnableTUI)

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
	case EnvDev, EnvProd:
	default:
		return nil, fmt.Errorf("invalid env: %s (must be dev or prod)", cfg.Env)
	}

	switch cfg.OutputType {
	case OutputTypePackdir, OutputTypeArchive:
	default:
		return nil, fmt.Errorf("invalid output_type: %s (must be packdir or archive)", cfg.OutputType)
	}

	return &cfg, nil
}

func (c *Config) IsDev() bool {
	return c.Env == EnvDev
}

func (c *Config) IsProd() bool {
	return c.Env == EnvProd
}
