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

	// Report generation errors, if any
	if generateErr != nil {
		output.Warning("Pack generation completed with errors: %v", generateErr)
		slog.Error("pack generation completed with errors", "error", generateErr)
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

func (w *Watcher) fetchServers(ctx context.Context) ([]v0.ServerJSON, error) {
	// Fetch by name if name filter provided
	if len(w.config.NameFilter.Names) > 0 {
		return w.fetchServersByName(ctx)
	}

	// Fetch all servers if no name filter provided
	opts := &mcp.ServerListOptions{}
	allServers, err := w.client.Servers.ListAll(ctx, opts)
	if err != nil {
		return nil, err
	}

	return allServers, nil
}

func (w *Watcher) fetchServersByName(ctx context.Context) ([]v0.ServerJSON, error) {
	var allServers []v0.ServerJSON

	// Track seen servers to manage deduplication
	seenServers := make(map[string]bool)

	// Fetch servers by exact name for each name filter
	for _, nameFilter := range w.config.NameFilter.Names {
		slog.Debug("fetching servers by name", "name", nameFilter)

		servers, err := w.client.Servers.GetByName(ctx, nameFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to get servers by name %s: %w", nameFilter, err)
		}

		// Deduplicate server entries in case user provided duplicate name filters
		for _, server := range servers {
			serverKey := fmt.Sprintf("%s@%s", server.Name, server.Version)
			if !seenServers[serverKey] {
				allServers = append(allServers, server)
				seenServers[serverKey] = true
			}
		}
	}

	slog.Debug("fetch by name completed", "total_servers", len(allServers))

	return allServers, nil
}

func (w *Watcher) filterServers(servers []v0.ServerJSON) []ServerGenerateTask {
	var tasks []ServerGenerateTask

	for _, srv := range servers {
		nameSpec, err := server.ParseNameSpec(srv.Name)
		if err != nil {
			output.Warning("Polled server %q has invalid server name format, skipping", srv.Name)
			slog.Warn("invalid server name format found during watcher filtering", "name", srv.Name, "error", err)
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

			// Check if generation is needed based on state
			if w.state.NeedsGeneration(namespace, name, srv.Version, pkg.RegistryType, pkg.Transport.Type, time.Time{}) {
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

	wg.Wait()
	close(failureChan)
	close(successChan)

	var errs []error
	var criticalErrs []error
	var successCount int

	for err := range failureChan {
		errs = append(errs, err)
		output.Failure("Pack generation failed: %v", err)
		slog.Error("pack generation failed", "error", err)

		if !errors.Is(err, generator.ErrPackDirectoryExists) &&
			!errors.Is(err, generator.ErrPackArchiveExists) {
			criticalErrs = append(criticalErrs, err)
		}
	}

	for range successChan {
		successCount++
	}

	if len(criticalErrs) > 0 {
		return successCount, fmt.Errorf("generation completed with %d critical errors", len(criticalErrs))
	}

	if len(errs) > 0 {
		slog.Info("pack generation completed with errors", "total_errors", len(errs), "critical_errors", len(criticalErrs))
	}

	return successCount, nil
}

func (w *Watcher) generatePack(ctx context.Context, task ServerGenerateTask) error {
	serverName := task.Server.Name

	output.Progress("Generating pack: %s@%s (%s, %s)", serverName, task.Server.Version, task.Package.RegistryType, task.Package.Transport.Type)
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

	output.Success("Pack generated: %s@%s (%s, %s)", serverName, task.Server.Version, task.Package.RegistryType, task.Package.Transport.Type)
	slog.Info("pack generated successfully",
		"server", serverName,
		"version", task.Server.Version,
		"package_type", task.Package.RegistryType,
		"transport_type", task.Package.Transport.Type,
	)

	return nil
}
