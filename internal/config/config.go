package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.nomad-mcp-pack/")

	viper.SetDefault("registry_url", DefaultConfig.RegistryURL)
	viper.SetDefault("log_level", DefaultConfig.LogLevel)
	viper.SetDefault("env", DefaultConfig.Env)
	viper.SetDefault("output_dir", DefaultConfig.OutputDir)
	viper.SetDefault("output_type", DefaultConfig.OutputType)
	viper.SetDefault("dry_run", DefaultConfig.DryRun)
	viper.SetDefault("force_overwrite", DefaultConfig.ForceOverwrite)
	viper.SetDefault("allow_deprecated", DefaultConfig.AllowDeprecated)
	viper.SetDefault("generate.package_type", DefaultConfig.GeneratePackageType)
	viper.SetDefault("generate.transport_type", DefaultConfig.GenerateTransportType)
	viper.SetDefault("server.addr", DefaultConfig.ServerAddr)
	viper.SetDefault("server.read_timeout", DefaultConfig.ServerReadTimeout)
	viper.SetDefault("server.write_timeout", DefaultConfig.ServerWriteTimeout)
	viper.SetDefault("watch.poll_interval", DefaultConfig.WatchPollInterval)
	viper.SetDefault("watch.filter_server_names", DefaultConfig.WatchFilterServerNames)
	viper.SetDefault("watch.filter_package_types", DefaultConfig.WatchFilterPackageTypes)
	viper.SetDefault("watch.filter_transport_types", DefaultConfig.WatchFilterTransportTypes)
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
