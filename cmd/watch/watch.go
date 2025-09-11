package cmdwatch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/watch"
	"github.com/leefowlercu/nomad-mcp-pack/pkg/generate"
	"github.com/leefowlercu/nomad-mcp-pack/pkg/registry"
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

	client, err := registry.NewClient(cfg.MCPRegistryURL)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	generateOpts := generate.Options{
		OutputDir:  cfg.OutputDir,
		OutputType: string(cfg.OutputType),
		DryRun:     dryRun,
		Force:      false,
	}

	watcher, err := watch.NewWatcher(cfg, client, generateOpts)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("received shutdown signal, stopping watch mode...")
		cancel()
	}()

	err = watcher.Run(ctx)
	if err != nil {
		if errors.Is(err, watch.ErrGracefulShutdown) {
			return nil
		}
		return err
	}
	return nil
}
