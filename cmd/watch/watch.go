package cmdwatch

import (
	"log/slog"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

var WatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Continuously poll the configured MCP Registry and generate Nomad Packs for new/updated MCP Servers",
	Long: "\nWatch the configured MCP Registry (default: https://registry.modelcontextprotocol.io) for new or updated MCP Servers and automatically generate Nomad Packs.\n\n" +
		"This command continuously polls the configured MCP Registry at the specified interval, tracks state to avoid regenerating unchanged packs, " +
		"and supports filtering options.",
	Example: `  # Watch all servers with default settings
  nomad-mcp-pack watch
  
  # Watch with custom output directory
  nomad-mcp-pack watch --output-dir ./generated-packs
  
  # Custom poll interval (in seconds)
  nomad-mcp-pack watch --poll-interval 300
  
  # Dry run to see what would be generated
  nomad-mcp-pack watch --dry-run`,
	RunE: runWatch,
}

func init() {
	WatchCmd.Flags().Int("poll-interval", config.DefaultConfig.WatchPollInterval, "Polling interval in seconds")
	WatchCmd.Flags().String("filter-names", config.DefaultConfig.WatchFilterNames, "Filter by MCP Server names (comma-separated values)")
	WatchCmd.Flags().String("filter-package-type", config.DefaultConfig.WatchFilterPackageTypes, "Filter by supported package types (comma-separated values)")
	WatchCmd.Flags().String("state-file", config.DefaultConfig.WatchStateFile, "Path to state file")
	WatchCmd.Flags().Int("max-concurrent", config.DefaultConfig.WatchMaxConcurrent, "Maximum concurrent pack generations")
	WatchCmd.Flags().Bool("enable-tui", config.DefaultConfig.WatchEnableTUI, "Show a Terminal UI instead of a log stream")
	WatchCmd.Flags().String("output-dir", config.DefaultConfig.OutputDir, "Output directory for generated packs")
	WatchCmd.Flags().String("output-type", config.DefaultConfig.OutputType, "Output type {packdir|archive}")
	WatchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be generated without writing files")
}

func runWatch(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	slog.Info("starting watch mode",
		"output_dir", cfg.OutputDir,
		"interval", cfg.Watch.PollInterval,
		"filter_names", cfg.Watch.FilterNames,
		"filter_package_types", cfg.Watch.FilterPackageTypes,
		"state_file", cfg.Watch.StateFile,
		"max_concurrent", cfg.Watch.MaxConcurrent,
		"enable_tui", cfg.Watch.EnableTUI,
		"dry_run", dryRun,
	)

	// TODO: Implement watch functionality
	// - Poll registry at interval
	// - Filter servers based on configuration
	// - Track state to avoid regenerating unchanged packs
	// - Generate packs for new/updated servers

	return nil
}
