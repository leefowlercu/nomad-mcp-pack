package watch

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestNewWatchState(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates valid watch state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewWatchState()
			if state == nil {
				t.Fatal("NewWatchState() returned nil")
			}
			if state.Servers == nil {
				t.Fatal("NewWatchState() Servers map not initialized")
			}
			if !state.GetLastPoll().IsZero() {
				t.Error("NewWatchState() initial last poll should be zero")
			}
		})
	}
}

func TestServerState_Key(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		serverName  string
		version     string
		packageType string
		want        string
	}{
		{
			name:        "standard key format",
			namespace:   "io.github.example",
			serverName:  "test-server",
			version:     "1.0.0",
			packageType: "npm",
			want:        "io.github.example/test-server@1.0.0:npm",
		},
		{
			name:        "complex namespace",
			namespace:   "org.example.sub",
			serverName:  "my-server",
			version:     "2.1.0-beta.1",
			packageType: "oci",
			want:        "org.example.sub/my-server@2.1.0-beta.1:oci",
		},
		{
			name:        "pypi package type",
			namespace:   "com.company",
			serverName:  "python-server",
			version:     "0.1.0",
			packageType: "pypi",
			want:        "com.company/python-server@0.1.0:pypi",
		},
		{
			name:        "nuget package type",
			namespace:   "io.nuget",
			serverName:  "dotnet-server",
			version:     "3.0.0",
			packageType: "nuget",
			want:        "io.nuget/dotnet-server@3.0.0:nuget",
		},
		{
			name:        "empty values",
			namespace:   "",
			serverName:  "",
			version:     "",
			packageType: "",
			want:        "/@:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ServerState{
				Namespace:   tt.namespace,
				Name:        tt.serverName,
				Version:     tt.version,
				PackageType: tt.packageType,
			}

			got := server.Key()
			if got != tt.want {
				t.Errorf("ServerState.Key() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWatchState_SetAndGetServer(t *testing.T) {
	baseTime := time.Now().Truncate(time.Second)

	tests := []struct {
		name   string
		server *ServerState
	}{
		{
			name: "npm server",
			server: &ServerState{
				Namespace:   "io.github.example",
				Name:        "npm-server",
				Version:     "1.0.0",
				PackageType: "npm",
				UpdatedAt:   baseTime,
				GeneratedAt: baseTime,
			},
		},
		{
			name: "oci server",
			server: &ServerState{
				Namespace:   "docker.io",
				Name:        "container-server",
				Version:     "2.1.0",
				PackageType: "oci",
				UpdatedAt:   baseTime.Add(time.Hour),
				GeneratedAt: baseTime.Add(time.Hour),
			},
		},
		{
			name: "pypi server",
			server: &ServerState{
				Namespace:   "pypi.org",
				Name:        "python-package",
				Version:     "0.5.0-alpha",
				PackageType: "pypi",
				UpdatedAt:   baseTime.Add(-time.Hour),
				GeneratedAt: baseTime.Add(-time.Hour),
				Checksum:    "abc123",
			},
		},
		{
			name: "server with empty checksum",
			server: &ServerState{
				Namespace:   "example.com",
				Name:        "test",
				Version:     "1.0.0",
				PackageType: "nuget",
				UpdatedAt:   baseTime,
				GeneratedAt: baseTime,
				Checksum:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewWatchState()
			
			// Set server
			state.SetServer(tt.server)
			
			// Get server
			retrieved, exists := state.GetServer(tt.server.Key())
			if !exists {
				t.Fatal("GetServer() server not found after setting")
			}

			// Verify all fields
			if retrieved.Namespace != tt.server.Namespace {
				t.Errorf("GetServer() namespace = %q, want %q", retrieved.Namespace, tt.server.Namespace)
			}
			if retrieved.Name != tt.server.Name {
				t.Errorf("GetServer() name = %q, want %q", retrieved.Name, tt.server.Name)
			}
			if retrieved.Version != tt.server.Version {
				t.Errorf("GetServer() version = %q, want %q", retrieved.Version, tt.server.Version)
			}
			if retrieved.PackageType != tt.server.PackageType {
				t.Errorf("GetServer() packageType = %q, want %q", retrieved.PackageType, tt.server.PackageType)
			}
			if !retrieved.UpdatedAt.Equal(tt.server.UpdatedAt) {
				t.Errorf("GetServer() updatedAt = %v, want %v", retrieved.UpdatedAt, tt.server.UpdatedAt)
			}
			if !retrieved.GeneratedAt.Equal(tt.server.GeneratedAt) {
				t.Errorf("GetServer() generatedAt = %v, want %v", retrieved.GeneratedAt, tt.server.GeneratedAt)
			}
			if retrieved.Checksum != tt.server.Checksum {
				t.Errorf("GetServer() checksum = %q, want %q", retrieved.Checksum, tt.server.Checksum)
			}
		})
	}
}

func TestWatchState_GetServer_NotFound(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{
			name: "nonexistent key",
			key:  "io.github.missing/server@1.0.0:npm",
		},
		{
			name: "empty key",
			key:  "",
		},
		{
			name: "invalid key format",
			key:  "invalid-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewWatchState()
			
			_, exists := state.GetServer(tt.key)
			if exists {
				t.Error("GetServer() should return false for nonexistent key")
			}
		})
	}
}

func TestWatchState_NeedsGeneration(t *testing.T) {
	baseTime := time.Now().Truncate(time.Second)

	tests := []struct {
		name        string
		setupServer *ServerState
		namespace   string
		serverName  string
		version     string
		packageType string
		updatedAt   time.Time
		want        bool
	}{
		{
			name:        "new server needs generation",
			setupServer: nil,
			namespace:   "io.github.new",
			serverName:  "server",
			version:     "1.0.0",
			packageType: "npm",
			updatedAt:   baseTime,
			want:        true,
		},
		{
			name: "server with same update time does not need generation",
			setupServer: &ServerState{
				Namespace:   "io.github.existing",
				Name:        "server",
				Version:     "1.0.0",
				PackageType: "npm",
				UpdatedAt:   baseTime,
				GeneratedAt: baseTime,
			},
			namespace:   "io.github.existing",
			serverName:  "server",
			version:     "1.0.0",
			packageType: "npm",
			updatedAt:   baseTime,
			want:        false,
		},
		{
			name: "server with older update time does not need generation",
			setupServer: &ServerState{
				Namespace:   "io.github.existing",
				Name:        "server",
				Version:     "1.0.0",
				PackageType: "npm",
				UpdatedAt:   baseTime,
				GeneratedAt: baseTime,
			},
			namespace:   "io.github.existing",
			serverName:  "server",
			version:     "1.0.0",
			packageType: "npm",
			updatedAt:   baseTime.Add(-time.Hour),
			want:        false,
		},
		{
			name: "server with newer update time needs generation",
			setupServer: &ServerState{
				Namespace:   "io.github.existing",
				Name:        "server",
				Version:     "1.0.0",
				PackageType: "npm",
				UpdatedAt:   baseTime,
				GeneratedAt: baseTime,
			},
			namespace:   "io.github.existing",
			serverName:  "server",
			version:     "1.0.0",
			packageType: "npm",
			updatedAt:   baseTime.Add(time.Hour),
			want:        true,
		},
		{
			name: "different version needs generation",
			setupServer: &ServerState{
				Namespace:   "io.github.existing",
				Name:        "server",
				Version:     "1.0.0",
				PackageType: "npm",
				UpdatedAt:   baseTime,
				GeneratedAt: baseTime,
			},
			namespace:   "io.github.existing",
			serverName:  "server",
			version:     "2.0.0",
			packageType: "npm",
			updatedAt:   baseTime,
			want:        true,
		},
		{
			name: "different package type needs generation",
			setupServer: &ServerState{
				Namespace:   "io.github.existing",
				Name:        "server",
				Version:     "1.0.0",
				PackageType: "npm",
				UpdatedAt:   baseTime,
				GeneratedAt: baseTime,
			},
			namespace:   "io.github.existing",
			serverName:  "server",
			version:     "1.0.0",
			packageType: "oci",
			updatedAt:   baseTime,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewWatchState()
			
			// Setup existing server if provided
			if tt.setupServer != nil {
				state.SetServer(tt.setupServer)
			}
			
			got := state.NeedsGeneration(tt.namespace, tt.serverName, tt.version, tt.packageType, tt.updatedAt)
			if got != tt.want {
				t.Errorf("NeedsGeneration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWatchState_LastPoll(t *testing.T) {
	baseTime := time.Now().Truncate(time.Second)

	tests := []struct {
		name      string
		pollTimes []time.Time
		want      time.Time
	}{
		{
			name:      "initial state has zero time",
			pollTimes: []time.Time{},
			want:      time.Time{},
		},
		{
			name:      "single poll time",
			pollTimes: []time.Time{baseTime},
			want:      baseTime,
		},
		{
			name: "multiple poll times return last set",
			pollTimes: []time.Time{
				baseTime.Add(-time.Hour),
				baseTime,
				baseTime.Add(-30*time.Minute),
			},
			want: baseTime.Add(-30*time.Minute),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewWatchState()
			
			// Update poll times
			for _, pollTime := range tt.pollTimes {
				state.UpdateLastPoll(pollTime)
			}
			
			got := state.GetLastPoll()
			if len(tt.pollTimes) == 0 {
				if !got.IsZero() {
					t.Errorf("GetLastPoll() = %v, want zero time", got)
				}
			} else {
				if !got.Equal(tt.want) {
					t.Errorf("GetLastPoll() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestWatchState_CleanupOldServers(t *testing.T) {
	baseTime := time.Now().Truncate(time.Second)

	tests := []struct {
		name          string
		servers       []*ServerState
		maxAge        time.Duration
		wantRemoved   int
		wantRemaining []string
	}{
		{
			name:          "empty state cleanup",
			servers:       []*ServerState{},
			maxAge:        24 * time.Hour,
			wantRemoved:   0,
			wantRemaining: []string{},
		},
		{
			name: "no old servers to cleanup",
			servers: []*ServerState{
				{
					Namespace:   "recent1",
					Name:        "server",
					Version:     "1.0.0",
					PackageType: "npm",
					UpdatedAt:   baseTime.Add(-1 * time.Hour),
					GeneratedAt: baseTime.Add(-1 * time.Hour),
				},
				{
					Namespace:   "recent2",
					Name:        "server",
					Version:     "1.0.0",
					PackageType: "oci",
					UpdatedAt:   baseTime.Add(-12 * time.Hour),
					GeneratedAt: baseTime.Add(-12 * time.Hour),
				},
			},
			maxAge:      24 * time.Hour,
			wantRemoved: 0,
			wantRemaining: []string{
				"recent1/server@1.0.0:npm",
				"recent2/server@1.0.0:oci",
			},
		},
		{
			name: "cleanup old servers",
			servers: []*ServerState{
				{
					Namespace:   "old1",
					Name:        "server",
					Version:     "1.0.0",
					PackageType: "npm",
					UpdatedAt:   baseTime.Add(-48 * time.Hour),
					GeneratedAt: baseTime.Add(-48 * time.Hour),
				},
				{
					Namespace:   "recent",
					Name:        "server",
					Version:     "1.0.0",
					PackageType: "oci",
					UpdatedAt:   baseTime.Add(-1 * time.Hour),
					GeneratedAt: baseTime.Add(-1 * time.Hour),
				},
				{
					Namespace:   "old2",
					Name:        "server",
					Version:     "2.0.0",
					PackageType: "pypi",
					UpdatedAt:   baseTime.Add(-72 * time.Hour),
					GeneratedAt: baseTime.Add(-72 * time.Hour),
				},
			},
			maxAge:      24 * time.Hour,
			wantRemoved: 2,
			wantRemaining: []string{
				"recent/server@1.0.0:oci",
			},
		},
		{
			name: "cleanup all servers",
			servers: []*ServerState{
				{
					Namespace:   "old1",
					Name:        "server",
					Version:     "1.0.0",
					PackageType: "npm",
					UpdatedAt:   baseTime.Add(-48 * time.Hour),
					GeneratedAt: baseTime.Add(-48 * time.Hour),
				},
				{
					Namespace:   "old2",
					Name:        "server",
					Version:     "1.0.0",
					PackageType: "oci",
					UpdatedAt:   baseTime.Add(-36 * time.Hour),
					GeneratedAt: baseTime.Add(-36 * time.Hour),
				},
			},
			maxAge:        24 * time.Hour,
			wantRemoved:   2,
			wantRemaining: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewWatchState()
			
			// Setup servers
			for _, server := range tt.servers {
				state.SetServer(server)
			}
			
			// Cleanup old servers
			removed := state.CleanupOldServers(tt.maxAge)
			if removed != tt.wantRemoved {
				t.Errorf("CleanupOldServers() removed = %d, want %d", removed, tt.wantRemoved)
			}
			
			// Verify remaining servers
			for _, expectedKey := range tt.wantRemaining {
				if _, exists := state.GetServer(expectedKey); !exists {
					t.Errorf("CleanupOldServers() expected server %q to remain", expectedKey)
				}
			}
			
			// Verify removed servers are gone
			for _, server := range tt.servers {
				key := server.Key()
				if !slices.Contains(tt.wantRemaining, key) {
					if _, exists := state.GetServer(key); exists {
						t.Errorf("CleanupOldServers() expected server %q to be removed", key)
					}
				}
			}
		})
	}
}

func TestLoadState(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func(t *testing.T) string
		want        func(*WatchState) bool
		expectError bool
		errorSubstr string
	}{
		{
			name: "nonexistent file returns empty state",
			setupFile: func(t *testing.T) string {
				return "/nonexistent/path/state.json"
			},
			want: func(state *WatchState) bool {
				return state != nil && len(state.Servers) == 0 && state.GetLastPoll().IsZero()
			},
			expectError: false,
		},
		{
			name: "valid state file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				statePath := filepath.Join(tmpDir, "valid-state.json")
				
				state := NewWatchState()
				baseTime := time.Now().Truncate(time.Second)
				state.UpdateLastPoll(baseTime)
				
				server := &ServerState{
					Namespace:   "io.github.test",
					Name:        "server",
					Version:     "1.0.0",
					PackageType: "npm",
					UpdatedAt:   baseTime,
					GeneratedAt: baseTime,
					Checksum:    "abc123",
				}
				state.SetServer(server)
				
				if err := state.SaveState(statePath); err != nil {
					t.Fatalf("failed to setup test file: %v", err)
				}
				
				return statePath
			},
			want: func(state *WatchState) bool {
				if state == nil || len(state.Servers) != 1 {
					return false
				}
				server, exists := state.GetServer("io.github.test/server@1.0.0:npm")
				return exists && server.Checksum == "abc123"
			},
			expectError: false,
		},
		{
			name: "invalid JSON file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				statePath := filepath.Join(tmpDir, "invalid.json")
				
				if err := os.WriteFile(statePath, []byte("{invalid json"), 0644); err != nil {
					t.Fatalf("failed to setup test file: %v", err)
				}
				
				return statePath
			},
			want:        nil,
			expectError: true,
			errorSubstr: "failed to unmarshal state",
		},
		{
			name: "empty JSON file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				statePath := filepath.Join(tmpDir, "empty.json")
				
				if err := os.WriteFile(statePath, []byte("{}"), 0644); err != nil {
					t.Fatalf("failed to setup test file: %v", err)
				}
				
				return statePath
			},
			want: func(state *WatchState) bool {
				return state != nil && len(state.Servers) == 0
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFile(t)
			
			state, err := LoadState(path)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("LoadState() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("LoadState() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("LoadState() unexpected error = %v", err)
				return
			}
			
			if tt.want != nil && !tt.want(state) {
				t.Error("LoadState() returned state does not match expectations")
			}
		})
	}
}

func TestWatchState_SaveState(t *testing.T) {
	tests := []struct {
		name        string
		setupState  func() *WatchState
		setupPath   func(t *testing.T) string
		expectError bool
		errorSubstr string
	}{
		{
			name: "save to valid path",
			setupState: func() *WatchState {
				state := NewWatchState()
				state.UpdateLastPoll(time.Now().Truncate(time.Second))
				
				server := &ServerState{
					Namespace:   "io.github.test",
					Name:        "server",
					Version:     "1.0.0",
					PackageType: "npm",
					UpdatedAt:   time.Now().Truncate(time.Second),
					GeneratedAt: time.Now().Truncate(time.Second),
				}
				state.SetServer(server)
				
				return state
			},
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "test-state.json")
			},
			expectError: false,
		},
		{
			name: "save to invalid path",
			setupState: func() *WatchState {
				return NewWatchState()
			},
			setupPath: func(t *testing.T) string {
				return "/invalid/path/that/does/not/exist/state.json"
			},
			expectError: true,
			errorSubstr: "failed to write state file",
		},
		{
			name: "empty state save",
			setupState: func() *WatchState {
				return NewWatchState()
			},
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "empty-state.json")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setupState()
			path := tt.setupPath(t)
			
			err := state.SaveState(path)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("SaveState() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("SaveState() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("SaveState() unexpected error = %v", err)
				return
			}
			
			// Verify file was created and can be loaded
			if _, err := os.Stat(path); err != nil {
				t.Errorf("SaveState() file was not created: %v", err)
			}
			
			// Verify roundtrip
			loaded, err := LoadState(path)
			if err != nil {
				t.Errorf("SaveState() created file cannot be loaded: %v", err)
			}
			
			if !loaded.GetLastPoll().Equal(state.GetLastPoll()) {
				t.Error("SaveState() roundtrip failed for last poll time")
			}
		})
	}
}

func TestSaveStateAtomicity(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "atomic save operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			statePath := filepath.Join(tmpDir, "atomic-test.json")
			
			// Create initial state
			state1 := NewWatchState()
			time1 := time.Now().Truncate(time.Second)
			state1.UpdateLastPoll(time1)
			
			if err := state1.SaveState(statePath); err != nil {
				t.Fatalf("SaveState() failed to save initial state: %v", err)
			}
			
			// Verify file exists
			if _, err := os.Stat(statePath); err != nil {
				t.Fatalf("SaveState() state file should exist: %v", err)
			}
			
			// Save again (should atomically replace)
			state2 := NewWatchState()
			time2 := time1.Add(time.Hour)
			state2.UpdateLastPoll(time2)
			
			if err := state2.SaveState(statePath); err != nil {
				t.Fatalf("SaveState() failed to save second state: %v", err)
			}
			
			// Load and verify it's the second state
			loaded, err := LoadState(statePath)
			if err != nil {
				t.Fatalf("LoadState() failed to load after atomic save: %v", err)
			}
			
			if !loaded.GetLastPoll().Equal(time2) {
				t.Errorf("SaveState() atomic operation failed: got %v, want %v", loaded.GetLastPoll(), time2)
			}
		})
	}
}