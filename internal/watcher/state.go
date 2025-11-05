package watcher

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

type ServerState struct {
	Namespace     string    `json:"namespace"`
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	PackageType   string    `json:"package_type"`
	TransportType string    `json:"transport_type"`
	UpdatedAt     time.Time `json:"updated_at"`
	GeneratedAt   time.Time `json:"generated_at"`
	Checksum      string    `json:"checksum,omitempty"`
}

func (s *ServerState) Key() string {
	return fmt.Sprintf("%s/%s@%s:%s:%s", s.Namespace, s.Name, s.Version, s.PackageType, s.TransportType)
}

type WatchState struct {
	LastPoll time.Time               `json:"last_poll"`
	Servers  map[string]*ServerState `json:"servers"`
	mu       sync.RWMutex
}

func NewWatchState() *WatchState {
	return &WatchState{
		Servers: make(map[string]*ServerState),
	}
}

func LoadState(path string) (*WatchState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug("state file does not exist, starting with empty state", "path", path)
			return NewWatchState(), nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state WatchState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	if state.Servers == nil {
		state.Servers = make(map[string]*ServerState)
	}

	slog.Debug("state loaded from disk",
		"path", path,
		"servers_count", len(state.Servers),
		"last_poll", state.LastPoll,
	)

	return &state, nil
}

func (s *WatchState) SaveState(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	slog.Debug("saving state to disk",
		"path", path,
		"servers_count", len(s.Servers),
	)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temp file first to ensure atomic operation
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	// Move temp file to final location
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	slog.Debug("state saved successfully",
		"path", path,
		"servers_count", len(s.Servers),
	)

	return nil
}

func (s *WatchState) GetServer(key string) (*ServerState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	server, exists := s.Servers[key]
	return server, exists
}

func (s *WatchState) SetServer(server *ServerState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := server.Key()
	s.Servers[key] = server
	slog.Debug("state updated",
		"key", key,
		"namespace", server.Namespace,
		"name", server.Name,
		"version", server.Version,
		"package_type", server.PackageType,
		"transport_type", server.TransportType,
		"state_map_size", len(s.Servers),
	)
}

// TODO: Update logic to use some form of checksum or hash of the actual Pack data
func (s *WatchState) NeedsGeneration(namespace, name, version, packageType, transportType string, updatedAt time.Time) bool {
	key := fmt.Sprintf("%s/%s@%s:%s:%s", namespace, name, version, packageType, transportType)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check for existence of key
	existing, exists := s.Servers[key]
	if !exists {
		slog.Debug("state check: pack needs generation (key not in state)",
			"key", key,
			"state_map_size", len(s.Servers),
		)
		return true
	}

	// If server exists, but has a newer UpdatedAt, we need to regenerate
	needsRegen := updatedAt.After(existing.GeneratedAt)
	slog.Debug("state check: pack in state",
		"key", key,
		"needs_regeneration", needsRegen,
		"updated_at", updatedAt,
		"existing_generated_at", existing.GeneratedAt,
		"state_map_size", len(s.Servers),
	)
	return needsRegen
}

func (s *WatchState) UpdateLastPoll(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastPoll = t
}

func (s *WatchState) GetLastPoll() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LastPoll
}

func (s *WatchState) CleanupOldServers(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for key, server := range s.Servers {
		if server.UpdatedAt.Before(cutoff) {
			delete(s.Servers, key)
			removed++
		}
	}

	return removed
}
