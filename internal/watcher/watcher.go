package watcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/leefowlercu/go-mcp-registry/mcp"
	"github.com/leefowlercu/nomad-mcp-pack/internal/generator"
	"github.com/leefowlercu/nomad-mcp-pack/internal/output"
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
		"filter_server_names", w.config.NameFilter.Names,
		"filter_package_types", w.config.PackageFilter.Types,
		"filter_transport_types", w.config.TransportFilter.Types,
	)

	// Start the polling timer
	ticker := time.NewTicker(time.Duration(w.config.PollInterval) * time.Second)
	defer ticker.Stop()

	// Initial poll before entering the loop
	if err := w.poll(ctx); err != nil {
		slog.Error("initial poll failed", "error", err)
	}

	// Main polling loop, exits on context cancellation
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
	output.Progress("Starting poll cycle...")
	slog.Info("starting poll cycle", "start_time", startTime.Format(time.RFC3339))

	// Fetch servers from the registry, applying name filters if provided
	servers, err := w.fetchServers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch servers: %w", err)
	}

	output.Info("Fetched %d servers from registry", len(servers))
	slog.Info("watcher poll cycle; fetched servers from registry", "count", len(servers))

	// Figure out which servers need packs generated based on filters and state
	toGenerate := w.filterServers(servers)
	if len(toGenerate) == 0 {
		output.Info("No packs need generation")
		slog.Debug("no packs need generation")
		w.state.UpdateLastPoll(startTime)
		return w.state.SaveState(w.config.StateFilePath)
	}

	output.Info("%d packs need generation", len(toGenerate))
	slog.Info("watcher poll cycle; packs need generation", "count", len(toGenerate))

	// Generate packs
	successCount, generateErr := w.generatePacks(ctx, toGenerate)

	// Always update and save state, even if some generations failed
	w.state.UpdateLastPoll(startTime)
	if err := w.state.SaveState(w.config.StateFilePath); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// If there were critical errors during generation, log, wrap, and return them
	var packGenerationErrors *PackGenerationErrors
	if errors.As(generateErr, &packGenerationErrors) {
		for _, genErr := range packGenerationErrors.CriticalErrs {
			slog.Error("pack generation failed", "error", genErr)
		}

		return fmt.Errorf("failure during poll cycle; %w", packGenerationErrors)
	}

	// Report summary of the poll cycle
	if generateErr != nil {
		output.Info("Poll cycle completed (%v, %d succeeded, %d failed)", time.Since(startTime).Round(time.Second), successCount, len(toGenerate)-successCount)
	} else {
		output.Info("Poll cycle completed (%v, %d packs generated)", time.Since(startTime).Round(time.Second), successCount)
	}
	slog.Info("poll cycle completed", "duration", time.Since(startTime), "generated", successCount, "total_attempted", len(toGenerate))

	return nil
}

func (w *Watcher) fetchServers(ctx context.Context) ([]v0.ServerResponse, error) {
	// Fetch by name if name filter provided
	if len(w.config.NameFilter.Names) > 0 {
		return w.fetchServersByName(ctx)
	}

	// Fetch all servers if no name filter provided
	opts := &mcp.ServerListOptions{}
	return w.listAllServers(ctx, opts)
}

func (w *Watcher) fetchServersByName(ctx context.Context) ([]v0.ServerResponse, error) {
	var allServers []v0.ServerResponse

	// Track seen servers to manage deduplication
	seenServers := make(map[string]bool)

	// Fetch servers by exact name for each name filter
	for _, nameFilter := range w.config.NameFilter.Names {
		slog.Debug("fetching servers by name", "name", nameFilter)

		opts := &mcp.ServerListOptions{
			Search: nameFilter,
		}

		servers, err := w.listAllServers(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get servers by name %s: %w", nameFilter, err)
		}

		// Deduplicate server entries in case user provided duplicate name filters
		for _, serverResp := range servers {
			server := serverResp.Server
			serverKey := fmt.Sprintf("%s@%s", server.Name, server.Version)
			if !seenServers[serverKey] {
				allServers = append(allServers, serverResp)
				seenServers[serverKey] = true
			}
		}
	}

	slog.Debug("fetch by name completed", "total_servers", len(allServers))

	return allServers, nil
}

// listAllServers fetches all servers using pagination, returning ServerResponse objects
func (w *Watcher) listAllServers(ctx context.Context, opts *mcp.ServerListOptions) ([]v0.ServerResponse, error) {
	var allServers []v0.ServerResponse

	for {
		listResp, resp, err := w.client.Servers.List(ctx, opts)
		if err != nil {
			return nil, err
		}

		if listResp != nil {
			allServers = append(allServers, listResp.Servers...)
		}

		// Check if there are more pages
		if resp == nil || resp.NextCursor == "" {
			break
		}

		// Set up for next page
		if opts.ListOptions.Cursor == "" {
			opts.ListOptions.Cursor = resp.NextCursor
		} else {
			opts.Cursor = resp.NextCursor
		}
	}

	return allServers, nil
}

func (w *Watcher) filterServers(servers []v0.ServerResponse) []ServerGenerateTask {
	var tasks []ServerGenerateTask

	for _, serverResp := range servers {
		srv := serverResp.Server

		nameSpec, err := server.ParseNameSpec(srv.Name)
		if err != nil {
			slog.Warn("invalid server name format found during watcher filtering, skipping", "name", srv.Name, "error", err)
			continue
		}

		namespace := nameSpec.Namespace
		name := nameSpec.Name

		// Check against provided name filter
		if !w.config.NameFilter.Matches(srv.Name) {
			slog.Debug("polled server does not match name filter, skipping", "server", srv.Name)
			continue
		}

		// Skip remote-only servers
		if len(srv.Packages) == 0 {
			slog.Debug("polled server matched name filter but is remote-only (no packages defined), skipping",
				"server", srv.Name,
			)
			continue
		}

		// Check against provided package type and transport type filters
		// Then check if generation is needed based on state
		for _, pkg := range srv.Packages {
			// Check against provided package type filter
			if !w.config.PackageFilter.Matches(pkg.RegistryType) {
				slog.Debug("polled server does not match package type filter, skipping",
					"server", srv.Name,
					"type", pkg.RegistryType,
				)
				continue
			}

			// Check against provided transport type filter
			if !w.config.TransportFilter.Matches(pkg.Transport.Type) {
				slog.Debug("polled server does not match transport type filter, skipping",
					"server", srv.Name,
					"package_type", pkg.RegistryType,
					"transport_type", pkg.Transport.Type,
				)
				continue
			}

			// Check if generation is needed based on state (or if force-overwrite is enabled)
			if w.generateOpts.ForceOverwrite || w.state.NeedsGeneration(namespace, name, srv.Version, pkg.RegistryType, pkg.Transport.Type, time.Time{}) {
				pkgCopy := pkg
				tasks = append(tasks, ServerGenerateTask{
					Server:  srv,
					Package: &pkgCopy,
				})
				slog.Debug("server needs generation",
					"server", srv.Name,
					"version", srv.Version,
					"package_type", pkg.RegistryType,
					"transport_type", pkg.Transport.Type,
				)
			}
		}
	}

	return tasks
}

func (w *Watcher) generatePacks(ctx context.Context, tasks []ServerGenerateTask) (int, error) {
	// Use semaphore for concurrency control
	sem := newPackGenSemaphore(w.config.MaxConcurrent)
	var wg sync.WaitGroup

	// Channels to collect results from goroutines
	failureChan := make(chan error, len(tasks))
	successChan := make(chan string, len(tasks))

	// Start goroutines for each task
	for _, task := range tasks {
		wg.Add(1)
		go func(t ServerGenerateTask) {
			defer wg.Done()

			sem.Acquire()
			defer sem.Release()

			if ctx.Err() != nil {
				return
			}

			// Generate the pack and send result to appropriate channel
			if err := w.generatePack(ctx, t); err != nil {
				failureChan <- fmt.Errorf("failed to generate %s@%s:%s:%s; %w",
					t.Server.Name, t.Server.Version, t.Package.RegistryType, t.Package.Transport.Type, err)
			} else {
				successChan <- fmt.Sprintf("%s@%s:%s:%s", t.Server.Name, t.Server.Version, t.Package.RegistryType, t.Package.Transport.Type)
			}
		}(task)
	}

	// Wait for all goroutines to finish and then close result collection channels
	wg.Wait()
	close(failureChan)
	close(successChan)

	// Store non-critical and critical (unexpected) errors separately for failure reporting
	var nonCriticalErrs []error
	var criticalErrs []error

	for err := range failureChan {
		if errors.Is(err, generator.ErrPackDirectoryExists) ||
			errors.Is(err, generator.ErrPackArchiveExists) {
			nonCriticalErrs = append(nonCriticalErrs, err)
		} else {
			criticalErrs = append(criticalErrs, err)
		}
	}

	// Count successful generations for reporting
	var successCount int
	for range successChan {
		successCount++
	}

	// Determine if we return an error based on presence of critical errors
	if len(criticalErrs) > 0 {
		return successCount, &PackGenerationErrors{CriticalErrs: criticalErrs}
	}

	if len(nonCriticalErrs) > 0 {
		slog.Warn("pack generation completed with noncritical errors", "errors", len(nonCriticalErrs))
	}

	return successCount, nil
}

func (w *Watcher) generatePack(ctx context.Context, task ServerGenerateTask) error {
	serverName := task.Server.Name

	slog.Info("generating pack",
		"server", serverName,
		"version", task.Server.Version,
		"package_type", task.Package.RegistryType,
		"transport_type", task.Package.Transport.Type,
	)

	if err := generator.Run(ctx, &task.Server, task.Package, w.generateOpts); err != nil {
		return err
	}

	now := time.Now()
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
		PackageType:   task.Package.RegistryType,
		TransportType: task.Package.Transport.Type,
		UpdatedAt:     now,
		GeneratedAt:   now,
	}
	w.state.SetServer(state)

	slog.Info("pack generated successfully",
		"server", serverName,
		"version", task.Server.Version,
		"package_type", task.Package.RegistryType,
		"transport_type", task.Package.Transport.Type,
	)

	return nil
}
