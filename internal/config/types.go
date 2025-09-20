package config

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

type GenerateConfig struct {
	PackageType   string `mapstructure:"package_type"`
	TransportType string `mapstructure:"transport_type"`
}

type ServerConfig struct {
	Addr         string `mapstructure:"addr"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type WatchConfig struct {
	PollInterval         int      `mapstructure:"poll_interval"`
	FilterServerNames    []string `mapstructure:"filter_server_names"`
	FilterPackageTypes   []string `mapstructure:"filter_package_types"`
	FilterTransportTypes []string `mapstructure:"filter_transport_types"`
	StateFile            string   `mapstructure:"state_file"`
	MaxConcurrent        int      `mapstructure:"max_concurrent"`
	EnableTUI            bool     `mapstructure:"enable_tui"`
}

type Config struct {
	RegistryURL     string         `mapstructure:"registry_url"`
	LogLevel        LogLevel       `mapstructure:"log_level"`
	Env             Env            `mapstructure:"env"`
	OutputDir       string         `mapstructure:"output_dir"`
	OutputType      OutputType     `mapstructure:"output_type"`
	DryRun          bool           `mapstructure:"dry_run"`
	ForceOverwrite  bool           `mapstructure:"force_overwrite"`
	AllowDeprecated bool           `mapstructure:"allow_deprecated"`
	Silent          bool           `mapstructure:"silent"`
	Generate        GenerateConfig `mapstructure:"generate"`
	Server          ServerConfig   `mapstructure:"server"`
	Watch           WatchConfig    `mapstructure:"watch"`
}
