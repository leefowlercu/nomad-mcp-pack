package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/leefowlercu/go-mcp-registry/mcp"
	"github.com/leefowlercu/nomad-mcp-pack/internal/generator"
	"github.com/leefowlercu/nomad-mcp-pack/internal/server"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
)

func NewWatcher(client *mcp.Client, cfg *WatcherConfig, generateOpts generator.Options) (*Watcher, error) {
	state, err := LoadState(cfg.StateFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return &Watcher{
		client:       client,
		config:       cfg,
		state:        state,
		generateOpts: generateOpts,
	}, nil
}

func (w *Watcher) Run(ctx context.Context) error {
	slog.Info("starting watch mode",
		"poll_interval", w.config.PollInterval,
		"state_file", w.config.StateFilePath,
		"filter_names", w.config.NameFilter.Names,
		"filter_package_types", w.config.PackageFilter.Types,
		"filter_transport_types", w.config.TransportFilter.Types,
	)

	ticker := time.NewTicker(time.Duration(w.config.PollInterval) * time.Second)
	defer ticker.Stop()

	if err := w.poll(ctx); err != nil {
		slog.Error("initial poll failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("watch mode stopped")
			if ctx.Err() == context.Canceled {
				return ErrGracefulShutdown
			}
			return ctx.Err()
		case <-ticker.C:
			if err := w.poll(ctx); err != nil {
				slog.Error("poll failed", "error", err)
			}
		}
	}
}

func (w *Watcher) poll(ctx context.Context) error {
	startTime := time.Now()
	slog.Info("starting poll cycle", "start_time", startTime.Format(time.RFC3339))

	servers, err := w.fetchServers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch servers: %w", err)
	}

	slog.Info("fetched servers from registry", "count", len(servers))

	toGenerate := w.filterServers(servers)
	if len(toGenerate) == 0 {
		slog.Debug("no servers need generation")
		w.state.UpdateLastPoll(startTime)
		return w.state.SaveState(w.config.StateFilePath)
	}

	slog.Info("servers need generation", "count", len(toGenerate))

	// Generate packs with concurrency control
	generateErr := w.generatePacks(ctx, toGenerate)

	// Always update and save state, even if some generations failed
	w.state.UpdateLastPoll(startTime)
	if err := w.state.SaveState(w.config.StateFilePath); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Return generation error after saving state
	if generateErr != nil {
		slog.Error("pack generation completed with errors", "error", generateErr)
	}

	slog.Info("poll cycle completed",
		"duration", time.Since(startTime),
		"generated", len(toGenerate),
	)

	return nil
}

func (w *Watcher) fetchServers(ctx context.Context) ([]v0.ServerJSON, error) {
	// If name filters are provided, use search API for efficiency
	if len(w.config.NameFilter.Names) > 0 {
		return w.fetchServersWithSearch(ctx)
	}

	// Fetch all servers using the ListAll helper
	opts := &mcp.ServerListOptions{}
	allServers, err := w.client.Servers.ListAll(ctx, opts)
	if err != nil {
		return nil, err
	}

	return allServers, nil
}

func (w *Watcher) fetchServersWithSearch(ctx context.Context) ([]v0.ServerJSON, error) {
	var allServers []v0.ServerJSON
	seenServers := make(map[string]bool) // To deduplicate servers

	// Search for each name filter individually
	for _, nameFilter := range w.config.NameFilter.Names {
		slog.Debug("searching for servers", "filter", nameFilter)

		opts := &mcp.ServerListOptions{
			Search: nameFilter,
		}

		servers, err := w.client.Servers.ListAll(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to search servers for %s: %w", nameFilter, err)
		}

		// Add unique servers
		for _, server := range servers {
			serverKey := fmt.Sprintf("%s@%s", server.Name, server.Version)
			if !seenServers[serverKey] {
				allServers = append(allServers, server)
				seenServers[serverKey] = true
			}
		}
	}

	slog.Debug("search completed", "total_servers", len(allServers))
	return allServers, nil
}

// filterServers filters servers based on configuration and state
func (w *Watcher) filterServers(servers []v0.ServerJSON) []ServerGenerateTask {
	var tasks []ServerGenerateTask

	for _, srv := range servers {
		// Parse server name to get namespace and name
		// The server name should be in format "namespace/name"
		nameSpec, err := server.ParseNameSpec(srv.Name)
		if err != nil {
			slog.Warn("invalid server name format", "name", srv.Name, "error", err)
			continue
		}

		namespace := nameSpec.Namespace
		name := nameSpec.Name
		serverFullName := fmt.Sprintf("%s/%s", namespace, name)

		// Check name filter
		if !w.config.NameFilter.Matches(serverFullName) {
			slog.Debug("server filtered by name", "server", serverFullName)
			continue
		}

		// Check if server has packages (skip remote-only servers)
		if len(srv.Packages) == 0 {
			slog.Debug("server matched filter but is remote-only (no packages defined), skipping",
				"server", serverFullName)
			continue
		}

		// Check each package type
		for _, pkg := range srv.Packages {
			// Check package type filter
			if !w.config.PackageFilter.Matches(pkg.RegistryType) {
				slog.Debug("package filtered by type",
					"server", serverFullName,
					"type", pkg.RegistryType,
				)
				continue
			}

			// Check transport type filter
			if !w.config.TransportFilter.Matches(pkg.Transport.Type) {
				slog.Debug("package filtered by transport type",
					"server", serverFullName,
					"package_type", pkg.RegistryType,
					"transport_type", pkg.Transport.Type,
				)
				continue
			}

			// Check if generation is needed based on state presence only
			// TODO: Add more sophisticated checks (e.g., checksum, updated_at) if available
			if w.state.NeedsGenerationWithTransport(namespace, name, srv.Version, pkg.RegistryType, pkg.Transport.Type, time.Time{}) {
				pkgCopy := pkg // Copy to avoid reference issues
				tasks = append(tasks, ServerGenerateTask{
					Server:        srv,
					PackageType:   pkg.RegistryType,
					TransportType: pkg.Transport.Type,
					Package:       &pkgCopy,
				})
				slog.Debug("server needs generation",
					"server", serverFullName,
					"version", srv.Version,
					"package_type", pkg.RegistryType,
					"transport_type", pkg.Transport.Type,
				)
			}
		}
	}

	return tasks
}

func (w *Watcher) generatePacks(ctx context.Context, tasks []ServerGenerateTask) error {
	sem := make(chan struct{}, w.config.MaxConcurrent)

	var wg sync.WaitGroup

	// Channel for errors
	errChan := make(chan error, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(t ServerGenerateTask) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				return
			}

			if err := w.generateSinglePack(ctx, t); err != nil {
				errChan <- fmt.Errorf("failed to generate %s@%s:%s: %w",
					t.Server.Name, t.Server.Version, t.PackageType, err)
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	var criticalErrs []error

	for err := range errChan {
		errs = append(errs, err)
		slog.Error("pack generation failed", "error", err)

		// Check if this is a critical error (not just "directory already exists")
		if !strings.Contains(err.Error(), "already exists") {
			criticalErrs = append(criticalErrs, err)
		}
	}

	// Only return error for critical failures, not "already exists" errors
	if len(criticalErrs) > 0 {
		return fmt.Errorf("generation completed with %d critical errors", len(criticalErrs))
	}

	// Log summary of all errors (including non-critical)
	if len(errs) > 0 {
		slog.Info("pack generation completed with non-critical errors",
			"total_errors", len(errs),
			"critical_errors", len(criticalErrs))
	}

	return nil
}

func (w *Watcher) generateSinglePack(ctx context.Context, task ServerGenerateTask) error {
	serverName := task.Server.Name

	slog.Info("generating pack",
		"server", serverName,
		"version", task.Server.Version,
		"package_type", task.PackageType,
		"transport_type", task.TransportType,
	)

	if err := generator.Run(ctx, &task.Server, task.Package, w.generateOpts); err != nil {
		return err
	}

	now := time.Now()
	// Parse namespace and name from server.Name
	nameSpec, err := server.ParseNameSpec(task.Server.Name)
	if err != nil {
		return fmt.Errorf("failed to parse server name: %w", err)
	}
	namespace := nameSpec.Namespace
	name := nameSpec.Name

	state := &ServerState{
		Namespace:     namespace,
		Name:          name,
		Version:       task.Server.Version,
		PackageType:   task.PackageType,
		TransportType: task.TransportType,
		UpdatedAt:     now,
		GeneratedAt:   now,
	}
	w.state.SetServer(state)

	slog.Info("pack generated successfully",
		"server", serverName,
		"version", task.Server.Version,
		"package_type", task.PackageType,
		"transport_type", task.TransportType,
	)

	return nil
}
