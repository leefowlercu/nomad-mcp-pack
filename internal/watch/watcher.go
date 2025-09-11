package watch

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/serversearchutils"
	"github.com/leefowlercu/nomad-mcp-pack/internal/watchutils"
	"github.com/leefowlercu/nomad-mcp-pack/pkg/generate"
	"github.com/leefowlercu/nomad-mcp-pack/pkg/registry"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

type Watcher struct {
	client        *registry.Client
	config        *config.Config
	nameFilter    *watchutils.ServerNameFilter
	packageFilter *watchutils.PackageTypeFilter
	state         *WatchState
	generateOpts  generate.Options
	logger        *slog.Logger
}

func NewWatcher(cfg *config.Config, client *registry.Client, generateOpts generate.Options) (*Watcher, error) {
	filterNames, err := watchutils.ParseFilterNames(cfg.Watch.FilterNames)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filter names: %w", err)
	}

	filterTypes, err := watchutils.ParsePackageTypes(cfg.Watch.FilterPackageTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package types: %w", err)
	}

	if err := watchutils.ValidateWatchConfig(
		cfg.Watch.PollInterval,
		cfg.Watch.StateFile,
		cfg.Watch.MaxConcurrent,
	); err != nil {
		return nil, fmt.Errorf("invalid watch configuration: %w", err)
	}

	state, err := LoadState(cfg.Watch.StateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return &Watcher{
		client:        client,
		config:        cfg,
		nameFilter:    &watchutils.ServerNameFilter{Names: filterNames},
		packageFilter: &watchutils.PackageTypeFilter{Types: filterTypes},
		state:         state,
		generateOpts:  generateOpts,
		logger:        slog.Default(),
	}, nil
}

func (w *Watcher) Run(ctx context.Context) error {
	w.logger.Info("starting watch mode",
		"poll_interval", w.config.Watch.PollInterval,
		"state_file", w.config.Watch.StateFile,
		"filter_names", w.config.Watch.FilterNames,
		"filter_types", w.config.Watch.FilterPackageTypes,
	)

	ticker := time.NewTicker(time.Duration(w.config.Watch.PollInterval) * time.Second)
	defer ticker.Stop()

	if err := w.poll(ctx); err != nil {
		w.logger.Error("initial poll failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("watch mode stopped")
			if ctx.Err() == context.Canceled {
				return ErrGracefulShutdown
			}
			return ctx.Err()
		case <-ticker.C:
			if err := w.poll(ctx); err != nil {
				w.logger.Error("poll failed", "error", err)
			}
		}
	}
}

func (w *Watcher) poll(ctx context.Context) error {
	startTime := time.Now()
	w.logger.Debug("starting poll cycle")

	lastPoll := w.state.GetLastPoll()
	var updatedSince string
	if !lastPoll.IsZero() {
		updatedSince = lastPoll.Format(time.RFC3339)
		w.logger.Debug("using updated_since filter", "timestamp", updatedSince)
	}

	servers, err := w.fetchServers(ctx, updatedSince)
	if err != nil {
		return fmt.Errorf("failed to fetch servers: %w", err)
	}

	w.logger.Info("fetched servers from registry", "count", len(servers))

	toGenerate := w.filterServers(servers)
	if len(toGenerate) == 0 {
		w.logger.Debug("no servers need generation")
		w.state.UpdateLastPoll(startTime)
		return w.state.SaveState(w.config.Watch.StateFile)
	}

	w.logger.Info("servers need generation", "count", len(toGenerate))

	// Generate packs with concurrency control
	if err := w.generatePacks(ctx, toGenerate); err != nil {
		return fmt.Errorf("pack generation failed: %w", err)
	}

	w.state.UpdateLastPoll(startTime)
	if err := w.state.SaveState(w.config.Watch.StateFile); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	w.logger.Info("poll cycle completed",
		"duration", time.Since(startTime),
		"generated", len(toGenerate),
	)

	return nil
}

// fetchServers retrieves servers from the registry
func (w *Watcher) fetchServers(ctx context.Context, updatedSince string) ([]v0.ServerJSON, error) {
	var allServers []v0.ServerJSON
	cursor := ""

	for {
		opts := &registry.ListServersOptions{
			Cursor:       cursor,
			Limit:        100,
			UpdatedSince: updatedSince,
		}

		resp, err := w.client.ListServers(ctx, opts)
		if err != nil {
			return nil, err
		}

		allServers = append(allServers, resp.Servers...)

		// Check for next page
		if resp.Metadata == nil || resp.Metadata.NextCursor == "" {
			break
		}
		cursor = resp.Metadata.NextCursor
	}

	return allServers, nil
}

// ServerGenerateTask represents a server and package type to generate
type ServerGenerateTask struct {
	Server      v0.ServerJSON
	ServerSpec  *serversearchutils.ServerSearchSpec
	PackageType string
	Package     *model.Package
}

// filterServers filters servers based on configuration and state
func (w *Watcher) filterServers(servers []v0.ServerJSON) []ServerGenerateTask {
	var tasks []ServerGenerateTask

	for _, server := range servers {
		// Parse server name to get namespace and name
		// The server name should be in format "namespace/name"
		namespace, name, err := serversearchutils.ValidateServerName(server.Name)
		if err != nil {
			w.logger.Warn("invalid server name format", "name", server.Name, "error", err)
			continue
		}

		serverFullName := fmt.Sprintf("%s/%s", namespace, name)

		// Check name filter
		if !w.nameFilter.Matches(serverFullName) {
			w.logger.Debug("server filtered by name", "server", serverFullName)
			continue
		}

		// Check each package type
		for _, pkg := range server.Packages {
			// Check package type filter
			if !w.packageFilter.Matches(pkg.RegistryType) {
				w.logger.Debug("package filtered by type",
					"server", serverFullName,
					"type", pkg.RegistryType,
				)
				continue
			}

			// Check if generation is needed
			updatedAt := time.Now() // Use current time as proxy since API doesn't provide updated timestamp

			if w.state.NeedsGeneration(namespace, name, server.Version, pkg.RegistryType, updatedAt) {
				pkgCopy := pkg // Copy to avoid reference issues
				tasks = append(tasks, ServerGenerateTask{
					Server: server,
					ServerSpec: &serversearchutils.ServerSearchSpec{
						Namespace: namespace,
						Name:      name,
						Version:   server.Version,
					},
					PackageType: pkg.RegistryType,
					Package:     &pkgCopy,
				})
				w.logger.Debug("server needs generation",
					"server", serverFullName,
					"version", server.Version,
					"type", pkg.RegistryType,
				)
			}
		}
	}

	return tasks
}

func (w *Watcher) generatePacks(ctx context.Context, tasks []ServerGenerateTask) error {
	sem := make(chan struct{}, w.config.Watch.MaxConcurrent)

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
					t.ServerSpec.FullName(), t.Server.Version, t.PackageType, err)
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
		w.logger.Error("pack generation failed", "error", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("generation completed with %d errors", len(errs))
	}

	return nil
}

func (w *Watcher) generateSinglePack(ctx context.Context, task ServerGenerateTask) error {
	serverName := task.ServerSpec.FullName()

	w.logger.Info("generating pack",
		"server", serverName,
		"version", task.Server.Version,
		"type", task.PackageType,
	)

	if err := generate.Run(ctx, &task.Server, task.ServerSpec, task.PackageType, w.generateOpts); err != nil {
		return err
	}

	now := time.Now()
	state := &ServerState{
		Namespace:   task.ServerSpec.Namespace,
		Name:        task.ServerSpec.Name,
		Version:     task.Server.Version,
		PackageType: task.PackageType,
		UpdatedAt:   now,
		GeneratedAt: now,
	}
	w.state.SetServer(state)

	if err := w.state.SaveState(w.config.Watch.StateFile); err != nil {
		w.logger.Error("failed to save state after generation", "error", err)
	}

	w.logger.Info("pack generated successfully",
		"server", serverName,
		"version", task.Server.Version,
		"type", task.PackageType,
	)

	return nil
}
