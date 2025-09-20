package cmdwatch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/leefowlercu/go-mcp-registry/mcp"
	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/generator"
	"github.com/leefowlercu/nomad-mcp-pack/internal/utils"
	"github.com/leefowlercu/nomad-mcp-pack/internal/validate"
	"github.com/leefowlercu/nomad-mcp-pack/internal/watcher"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

  # Filter by specific server names
  nomad-mcp-pack watch --filter-server-names "io.github.containers/kubernetes-mcp-server,ai.waystation/gmail"

  # Dry run to see what would be generated
  nomad-mcp-pack watch --dry-run`,
	PreRunE: runValidate,
	RunE:    runWatch,
}

func init() {
	WatchCmd.Flags().StringSlice("filter-server-names", config.DefaultConfig.WatchFilterServerNames, "Filter by MCP Server names (comma-separated values)")
	WatchCmd.Flags().StringSlice("filter-package-types", config.DefaultConfig.WatchFilterPackageTypes, "Filter by supported package types (comma-separated values)")
	WatchCmd.Flags().StringSlice("filter-transport-types", config.DefaultConfig.WatchFilterTransportTypes, "Filter by transport types (comma-separated values)")
	WatchCmd.Flags().Int("poll-interval", config.DefaultConfig.WatchPollInterval, "Polling interval in seconds")
	WatchCmd.Flags().String("state-file", config.DefaultConfig.WatchStateFile, "Path to state file")
	WatchCmd.Flags().Int("max-concurrent", config.DefaultConfig.WatchMaxConcurrent, "Maximum concurrent pack generations")
	WatchCmd.Flags().Bool("enable-tui", config.DefaultConfig.WatchEnableTUI, "Show a Terminal UI instead of a log stream")

	viper.BindPFlag("watch.filter_server_names", WatchCmd.Flags().Lookup("filter-server-names"))
	viper.BindPFlag("watch.filter_package_types", WatchCmd.Flags().Lookup("filter-package-types"))
	viper.BindPFlag("watch.filter_transport_types", WatchCmd.Flags().Lookup("filter-transport-types"))
	viper.BindPFlag("watch.poll_interval", WatchCmd.Flags().Lookup("poll-interval"))
	viper.BindPFlag("watch.state_file", WatchCmd.Flags().Lookup("state-file"))
	viper.BindPFlag("watch.max_concurrent", WatchCmd.Flags().Lookup("max-concurrent"))
	viper.BindPFlag("watch.enable_tui", WatchCmd.Flags().Lookup("enable-tui"))

	WatchCmd.Flags().SortFlags = false
}

func runValidate(cmd *cobra.Command, args []string) error {
	slog.Info("starting watch command input validation")

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration; %w", err)
	}

	slog.Debug("validating watch command inputs with configuration",
		slog.Group("common_config",
			"registry_url", cfg.RegistryURL,
			"log_level", cfg.LogLevel,
			"env", cfg.Env,
			"output_dir", cfg.OutputDir,
			"output_type", cfg.OutputType,
			"allow_deprecated", cfg.AllowDeprecated,
			"dry_run", cfg.DryRun,
			"force_overwrite", cfg.ForceOverwrite,
		),
		slog.Group("watch_config",
			"filter_server_names", cfg.Watch.FilterServerNames,
			"filter_package_types", cfg.Watch.FilterPackageTypes,
			"filter_transport_types", cfg.Watch.FilterTransportTypes,
			"poll_interval", cfg.Watch.PollInterval,
			"state_file", cfg.Watch.StateFile,
			"max_concurrent", cfg.Watch.MaxConcurrent,
			"enable_tui", cfg.Watch.EnableTUI,
		),
	)

	filterServerNames := cfg.Watch.FilterServerNames
	filterPackageTypes := cfg.Watch.FilterPackageTypes
	filterTransportTypes := cfg.Watch.FilterTransportTypes
	pollInterval := cfg.Watch.PollInterval
	stateFile := cfg.Watch.StateFile
	maxConcurrent := cfg.Watch.MaxConcurrent

	if err := validate.ServerNames(filterServerNames); err != nil {
		return fmt.Errorf("could not validate names filter; %w", err)
	}

	if err := validate.PackageTypes(filterPackageTypes, true); err != nil {
		return fmt.Errorf("could not validate package types filter; %w", err)
	}

	if err := validate.TransportTypes(filterTransportTypes, true); err != nil {
		return fmt.Errorf("could not validate transport types filter; %w", err)
	}

	if err := validate.PollInterval(pollInterval); err != nil {
		return fmt.Errorf("could not validate poll interval; %w", err)
	}

	if err := validate.StateFile(stateFile); err != nil {
		return fmt.Errorf("could not validate state file; %w", err)
	}

	if err := validate.MaxConcurrent(maxConcurrent); err != nil {
		return fmt.Errorf("could not validate max concurrent; %w", err)
	}

	slog.Info("watch command input validation completed successfully")

	// Any errors after this point are runtime errors, not usage-related errors
	cmd.SilenceUsage = true

	return nil
}

func runWatch(cmd *cobra.Command, args []string) error {
	slog.Info("starting watch command run")

	ctx := cmd.Context()

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration; %w", err)
	}

	slog.Debug("running watch command with configuration",
		slog.Group("common_config",
			"registry_url", cfg.RegistryURL,
			"log_level", cfg.LogLevel,
			"env", cfg.Env,
			"output_dir", cfg.OutputDir,
			"output_type", cfg.OutputType,
			"allow_deprecated", cfg.AllowDeprecated,
			"dry_run", cfg.DryRun,
			"force_overwrite", cfg.ForceOverwrite,
		),
		slog.Group("watch_config",
			"filter_server_names", cfg.Watch.FilterServerNames,
			"filter_package_types", cfg.Watch.FilterPackageTypes,
			"filter_transport_types", cfg.Watch.FilterTransportTypes,
			"poll_interval", cfg.Watch.PollInterval,
			"state_file", cfg.Watch.StateFile,
			"max_concurrent", cfg.Watch.MaxConcurrent,
			"enable_tui", cfg.Watch.EnableTUI,
		),
	)

	filterNames := cfg.Watch.FilterServerNames
	filterPackageTypes := cfg.Watch.FilterPackageTypes
	filterTransportTypes := cfg.Watch.FilterTransportTypes
	pollInterval := cfg.Watch.PollInterval
	stateFile := cfg.Watch.StateFile
	maxConcurrent := cfg.Watch.MaxConcurrent
	// enableTUI := cfg.Watch.EnableTUI

	registryURL := cfg.RegistryURL
	outputDir := cfg.OutputDir
	outputType := cfg.OutputType
	allowDeprecated := cfg.AllowDeprecated
	dryRun := cfg.DryRun
	forceOverwrite := cfg.ForceOverwrite

	client := mcp.NewClient(nil)
	registryURLParsed, err := url.Parse(registryURL)
	if err != nil {
		return fmt.Errorf("could not parse registry URL; %w", err)
	}
	client.BaseURL = registryURLParsed

	generateOpts := generator.Options{
		OutputDir:      outputDir,
		OutputType:     string(outputType),
		DryRun:         dryRun,
		ForceOverwrite: forceOverwrite,
	}

	watcherConfig := &watcher.WatcherConfig{
		PollInterval:    pollInterval,
		StateFilePath:   stateFile,
		MaxConcurrent:   maxConcurrent,
		AllowDeprecated: allowDeprecated,
		NameFilter: &watcher.ServerNameFilter{
			Names: utils.NormalizeAndDeduplicateStrings(filterNames),
		},
		PackageFilter: &watcher.PackageTypeFilter{
			Types: utils.NormalizeAndDeduplicateStrings(filterPackageTypes),
		},
		TransportFilter: &watcher.TransportTypeFilter{
			Types: utils.NormalizeAndDeduplicateStrings(filterTransportTypes),
		},
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle graceful shutdown on SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("received shutdown signal, stopping watch mode...")
		cancel()
	}()

	w, err := watcher.NewWatcher(client, watcherConfig, generateOpts)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	err = w.Run(ctx)
	if err != nil {
		if errors.Is(err, watcher.ErrGracefulShutdown) {
			return nil
		}
		return err
	}

	slog.Info("watch command run completed successfully")

	return nil
}
