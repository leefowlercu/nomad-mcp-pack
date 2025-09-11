package watch

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type ServerState struct {
	Namespace   string    `json:"namespace"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	PackageType string    `json:"package_type"`
	UpdatedAt   time.Time `json:"updated_at"`
	GeneratedAt time.Time `json:"generated_at"`
	Checksum    string    `json:"checksum,omitempty"`
}

func (s *ServerState) Key() string {
	return fmt.Sprintf("%s/%s@%s:%s", s.Namespace, s.Name, s.Version, s.PackageType)
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

	return &state, nil
}

func (s *WatchState) SaveState(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	s.Servers[server.Key()] = server
}

func (s *WatchState) NeedsGeneration(namespace, name, version, packageType string, updatedAt time.Time) bool {
	key := fmt.Sprintf("%s/%s@%s:%s", namespace, name, version, packageType)

	s.mu.RLock()
	defer s.mu.RUnlock()

	existing, exists := s.Servers[key]
	if !exists {
		return true
	}

	return updatedAt.After(existing.GeneratedAt)
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
