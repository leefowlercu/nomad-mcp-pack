package config

var ValidPackageTypes = []string{"npm", "pypi", "oci", "nuget"}

var ValidTransportTypes = []string{"stdio", "http", "sse"}

var ValidOutputTypes = []string{"packdir", "archive"}

const MinPollInterval = 30

const MinMaxConcurrent = 1

var DefaultConfig = struct {
	RegistryURL               string
	LogLevel                  string
	Env                       string
	OutputDir                 string
	OutputType                string
	DryRun                    bool
	ForceOverwrite            bool
	AllowDeprecated           bool
	Silent                    bool
	GeneratePackageType       string
	GenerateTransportType     string
	ServerAddr                string
	ServerReadTimeout         int
	ServerWriteTimeout        int
	WatchPollInterval         int
	WatchFilterServerNames    []string
	WatchFilterPackageTypes   []string
	WatchFilterTransportTypes []string
	WatchStateFile            string
	WatchMaxConcurrent        int
	WatchEnableTUI            bool
}{
	RegistryURL:               "https://registry.modelcontextprotocol.io/",
	LogLevel:                  "info",
	Env:                       "prod",
	OutputDir:                 "./packs",
	OutputType:                "packdir",
	DryRun:                    false,
	ForceOverwrite:            false,
	AllowDeprecated:           false,
	Silent:                    false,
	GeneratePackageType:       "oci",
	GenerateTransportType:     "http",
	ServerAddr:                ":8080",
	ServerReadTimeout:         10,
	ServerWriteTimeout:        10,
	WatchPollInterval:         300,
	WatchFilterServerNames:    []string{},
	WatchFilterPackageTypes:   ValidPackageTypes,
	WatchFilterTransportTypes: ValidTransportTypes,
	WatchStateFile:            "./watch.json",
	WatchMaxConcurrent:        5,
	WatchEnableTUI:            false,
}
