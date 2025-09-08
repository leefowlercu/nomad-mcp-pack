package cmdwatch

import (
	"log/slog"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dryRun    bool
	enableTUI bool
)

var WatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Continuously poll the configured MCP Registry and generate Nomad Packs for new/updated MCP Servers",
	Long: "Watch the configured MCP Registry (default: https://registry.modelcontextprotocol.io) for new or updated MCP Servers and automatically generate Nomad Packs.\n\n" +
		"This command continuously polls the configured MCP Registry at the specified interval, tracks state to avoid regenerating unchanged packs, " +
		"and supports filtering options.",
	Example: `  # Watch all servers with default settings
  nomad-mcp-pack watch
  
  # Watch with custom output directory
  nomad-mcp-pack watch --output ./generated-packs
  
  # Custom poll interval (in seconds)
  nomad-mcp-pack watch --interval 300
  
  # Dry run to see what would be generated
  nomad-mcp-pack watch --dry-run`,
	RunE: runWatch,
}

func init() {
	WatchCmd.Flags().StringP("output-dir", "o", "", "Output directory for generated packs")
	WatchCmd.Flags().Int("poll-interval", 0, "Polling interval in seconds")
	WatchCmd.Flags().String("filter-names", "", "Filter by MCP Server names (comma-separated values)")
	WatchCmd.Flags().String("filter-package-type", "", "Filter by supported package types (comma-separated values)")
	WatchCmd.Flags().String("state-file", "", "Path to state file")
	WatchCmd.Flags().Int("max-concurrent", 0, "Maximum concurrent pack generations")
	WatchCmd.Flags().Bool("enable-tui", false, "Show a Terminal UI instead of a log stream")
	WatchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be generated without writing files")

	viper.BindPFlag("output_dir", WatchCmd.Flags().Lookup("output-dir"))
	viper.BindPFlag("watch.poll_interval", WatchCmd.Flags().Lookup("poll-interval"))
	viper.BindPFlag("watch.filter_names", WatchCmd.Flags().Lookup("filter-names"))
	viper.BindPFlag("watch.filter_package_types", WatchCmd.Flags().Lookup("filter-package-type"))
	viper.BindPFlag("watch.state_file", WatchCmd.Flags().Lookup("state-file"))
	viper.BindPFlag("watch.max_concurrent", WatchCmd.Flags().Lookup("max-concurrent"))
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
