package cmd

import (
	"fmt"
	"log/slog"

	cmdgenerate "github.com/leefowlercu/nomad-mcp-pack/cmd/generate"
	cmdserver "github.com/leefowlercu/nomad-mcp-pack/cmd/server"
	cmdwatch "github.com/leefowlercu/nomad-mcp-pack/cmd/watch"
	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/logutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version string
)

var nomadMcpPackCmd = &cobra.Command{
	Use:   "nomad-mcp-pack",
	Short: "A generator for HashiCorp Nomad MCP Server Packs",
	Long: "\nNomad MCP Pack generates HashiCorp Nomad Packs for MCP Servers described " +
		"by the configured MCP Registry (default: https://registry.modelcontextprotocol.io)",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		config.InitConfig()

		// Bind command flags to configuration

		// Root command flags
		viper.BindPFlag("mcp_registry_url", cmd.Root().PersistentFlags().Lookup("mcp-registry-url"))

		// Generate command flags
		if outputDirFlag := cmd.Flag("output-dir"); outputDirFlag != nil {
			viper.BindPFlag("output_dir", outputDirFlag)
		}
		if outputTypeFlag := cmd.Flag("output-type"); outputTypeFlag != nil {
			viper.BindPFlag("output_type", outputTypeFlag)
		}

		// Server command flags
		if addrFlag := cmd.Flag("addr"); addrFlag != nil {
			viper.BindPFlag("server.addr", addrFlag)
		}
		if readTimeoutFlag := cmd.Flag("read-timeout"); readTimeoutFlag != nil {
			viper.BindPFlag("server.read_timeout", readTimeoutFlag)
		}
		if writeTimeoutFlag := cmd.Flag("write-timeout"); writeTimeoutFlag != nil {
			viper.BindPFlag("server.write_timeout", writeTimeoutFlag)
		}

		// Watch command flags
		if pollIntervalFlag := cmd.Flag("poll-interval"); pollIntervalFlag != nil {
			viper.BindPFlag("watch.poll_interval", pollIntervalFlag)
		}
		if filterNamesFlag := cmd.Flag("filter-names"); filterNamesFlag != nil {
			viper.BindPFlag("watch.filter_names", filterNamesFlag)
		}
		if filterPackageTypesFlag := cmd.Flag("filter-package-type"); filterPackageTypesFlag != nil {
			viper.BindPFlag("watch.filter_package_types", filterPackageTypesFlag)
		}
		if stateFileFlag := cmd.Flag("state-file"); stateFileFlag != nil {
			viper.BindPFlag("watch.state_file", stateFileFlag)
		}
		if maxConcurrentFlag := cmd.Flag("max-concurrent"); maxConcurrentFlag != nil {
			viper.BindPFlag("watch.max_concurrent", maxConcurrentFlag)
		}
		if enableTUIFlag := cmd.Flag("enable-tui"); enableTUIFlag != nil {
			viper.BindPFlag("watch.enable_tui", enableTUIFlag)
		}

		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Setup Logger
		logger := logutils.SetupLogger(cfg)
		slog.SetDefault(logger)

		slog.Info("starting nomad-mcp-pack",
			"mcp_registry_url", cfg.MCPRegistryURL,
			"log_level", cfg.LogLevel,
			"env", cfg.Env,
		)

		return nil
	},
}

func init() {
	nomadMcpPackCmd.PersistentFlags().String("mcp-registry-url", config.DefaultConfig.MCPRegistryURL, "MCP Registry URL")

	nomadMcpPackCmd.AddCommand(cmdgenerate.GenerateCmd)
	nomadMcpPackCmd.AddCommand(cmdserver.ServerCmd)
	nomadMcpPackCmd.AddCommand(cmdwatch.WatchCmd)
}

func Execute() error {
	return nomadMcpPackCmd.Execute()
}
