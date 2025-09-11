package watch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/pkg/generate"
	"github.com/leefowlercu/nomad-mcp-pack/pkg/registry"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

// Helper functions for creating test data
func createTestConfig(t *testing.T) *config.Config {
	tmpDir := t.TempDir()
	return &config.Config{
		MCPRegistryURL: "https://test-registry.example.com",
		OutputDir:      tmpDir,
		OutputType:     "packdir",
		Watch: config.WatchConfig{
			PollInterval:       60,
			FilterNames:        "io.github.test/server",
			FilterPackageTypes: "npm,oci",
			StateFile:          filepath.Join(tmpDir, "test-state.json"),
			MaxConcurrent:      2,
			EnableTUI:          false,
		},
	}
}

// createMockServer creates a mock HTTP server that returns predefined responses
func createMockServer(responses map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if response, exists := responses[path]; exists {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// createRegistryClient creates a registry client pointing to a test server
func createRegistryClient(t *testing.T, server *httptest.Server) *registry.Client {
	client, err := registry.NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create registry client: %v", err)
	}
	return client
}

func createTestServer(namespace, name, version string, packages []model.Package) v0.ServerJSON {
	return v0.ServerJSON{
		Name:     fmt.Sprintf("%s/%s", namespace, name),
		Version:  version,
		Packages: packages,
		Status:   model.StatusActive,
	}
}

func createTestPackage(registryType, identifier, version string) model.Package {
	return model.Package{
		RegistryType: registryType,
		Identifier:   identifier,
		Version:      version,
	}
}

func TestNewWatcher(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(t *testing.T) *config.Config
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid configuration creates watcher",
			setupConfig: func(t *testing.T) *config.Config {
				return createTestConfig(t)
			},
			expectError: false,
		},
		{
			name: "invalid filter names returns error",
			setupConfig: func(t *testing.T) *config.Config {
				cfg := createTestConfig(t)
				cfg.Watch.FilterNames = "invalid-name-format"
				return cfg
			},
			expectError: true,
			errorSubstr: "failed to parse filter names",
		},
		{
			name: "invalid package types returns error",
			setupConfig: func(t *testing.T) *config.Config {
				cfg := createTestConfig(t)
				cfg.Watch.FilterPackageTypes = "invalid,types"
				return cfg
			},
			expectError: true,
			errorSubstr: "failed to parse package types",
		},
		{
			name: "invalid poll interval returns error",
			setupConfig: func(t *testing.T) *config.Config {
				cfg := createTestConfig(t)
				cfg.Watch.PollInterval = 10 // Less than minimum of 30
				return cfg
			},
			expectError: true,
			errorSubstr: "invalid watch configuration",
		},
		{
			name: "empty state file returns error",
			setupConfig: func(t *testing.T) *config.Config {
				cfg := createTestConfig(t)
				cfg.Watch.StateFile = ""
				return cfg
			},
			expectError: true,
			errorSubstr: "invalid watch configuration",
		},
		{
			name: "invalid max concurrent returns error",
			setupConfig: func(t *testing.T) *config.Config {
				cfg := createTestConfig(t)
				cfg.Watch.MaxConcurrent = 0
				return cfg
			},
			expectError: true,
			errorSubstr: "invalid watch configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig(t)
			
			// Create a mock server for valid tests
			server := createMockServer(map[string]interface{}{
				"/v0/servers": v0.ServerListResponse{
					Servers: []v0.ServerJSON{},
					Metadata: &v0.Metadata{NextCursor: ""},
				},
			})
			defer server.Close()
			
			client := createRegistryClient(t, server)
			generateOpts := generate.Options{
				OutputDir:  cfg.OutputDir,
				OutputType: string(cfg.OutputType),
				DryRun:     true,
				Force:      false,
			}

			watcher, err := NewWatcher(cfg, client, generateOpts)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewWatcher() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("NewWatcher() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewWatcher() unexpected error = %v", err)
				return
			}

			if watcher == nil {
				t.Error("NewWatcher() returned nil watcher")
			}
		})
	}
}

func TestWatcher_fetchServers(t *testing.T) {
	tests := []struct {
		name         string
		setupServer  func() *httptest.Server
		updatedSince string
		want         int
		expectError  bool
		errorSubstr  string
	}{
		{
			name: "single page of results",
			setupServer: func() *httptest.Server {
				servers := []v0.ServerJSON{
					createTestServer("io.github.test", "server1", "1.0.0", []model.Package{
						createTestPackage("npm", "test-server1", "1.0.0"),
					}),
					createTestServer("io.github.test", "server2", "1.0.0", []model.Package{
						createTestPackage("oci", "test-server2", "1.0.0"),
					}),
				}
				return createMockServer(map[string]interface{}{
					"/v0/servers": v0.ServerListResponse{
						Servers: servers,
						Metadata: &v0.Metadata{
							NextCursor: "",
							Count:      2,
						},
					},
				})
			},
			updatedSince: "",
			want:         2,
			expectError:  false,
		},
		{
			name: "empty results",
			setupServer: func() *httptest.Server {
				return createMockServer(map[string]interface{}{
					"/v0/servers": v0.ServerListResponse{
						Servers: []v0.ServerJSON{},
						Metadata: &v0.Metadata{
							NextCursor: "",
							Count:      0,
						},
					},
				})
			},
			updatedSince: "",
			want:         0,
			expectError:  false,
		},
		{
			name: "registry error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}))
			},
			updatedSince: "",
			want:         0,
			expectError:  true,
			errorSubstr:  "unexpected status code 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(t)
			server := tt.setupServer()
			defer server.Close()
			
			client := createRegistryClient(t, server)
			watcher, err := NewWatcher(cfg, client, generate.Options{DryRun: true})
			if err != nil {
				t.Fatalf("Failed to create watcher: %v", err)
			}

			ctx := context.Background()
			servers, err := watcher.fetchServers(ctx, tt.updatedSince)

			if tt.expectError {
				if err == nil {
					t.Errorf("fetchServers() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("fetchServers() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("fetchServers() unexpected error = %v", err)
				return
			}

			if len(servers) != tt.want {
				t.Errorf("fetchServers() returned %d servers, want %d", len(servers), tt.want)
			}
		})
	}
}

func TestWatcher_filterServers(t *testing.T) {
	tests := []struct {
		name          string
		setupWatcher  func(t *testing.T) *Watcher
		servers       []v0.ServerJSON
		wantTaskCount int
	}{
		{
			name: "multiple package types per server",
			setupWatcher: func(t *testing.T) *Watcher {
				cfg := createTestConfig(t)
				cfg.Watch.FilterPackageTypes = "npm,oci"
				
				server := createMockServer(map[string]interface{}{
					"/v0/servers": v0.ServerListResponse{
						Servers: []v0.ServerJSON{},
						Metadata: &v0.Metadata{NextCursor: ""},
					},
				})
				defer server.Close()
				
				client := createRegistryClient(t, server)
				watcher, _ := NewWatcher(cfg, client, generate.Options{DryRun: true})
				return watcher
			},
			servers: []v0.ServerJSON{
				createTestServer("io.github.test", "server", "1.0.0", []model.Package{
					createTestPackage("npm", "test", "1.0.0"),
					createTestPackage("oci", "test", "1.0.0"),
					createTestPackage("pypi", "test", "1.0.0"), // Filtered out
				}),
			},
			wantTaskCount: 2,
		},
		{
			name: "invalid server name handling",
			setupWatcher: func(t *testing.T) *Watcher {
				cfg := createTestConfig(t)
				
				server := createMockServer(map[string]interface{}{
					"/v0/servers": v0.ServerListResponse{
						Servers: []v0.ServerJSON{},
						Metadata: &v0.Metadata{NextCursor: ""},
					},
				})
				defer server.Close()
				
				client := createRegistryClient(t, server)
				watcher, _ := NewWatcher(cfg, client, generate.Options{DryRun: true})
				return watcher
			},
			servers: []v0.ServerJSON{
				{
					Name:     "invalid-name", // Missing namespace
					Version:  "1.0.0",
					Packages: []model.Package{createTestPackage("npm", "test", "1.0.0")},
				},
			},
			wantTaskCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher := tt.setupWatcher(t)
			
			tasks := watcher.filterServers(tt.servers)
			
			if len(tasks) != tt.wantTaskCount {
				t.Errorf("filterServers() returned %d tasks, want %d", len(tasks), tt.wantTaskCount)
			}

			// Verify task structure
			for _, task := range tasks {
				if task.ServerSpec == nil {
					t.Error("filterServers() task missing ServerSpec")
				}
				if task.Package == nil {
					t.Error("filterServers() task missing Package")
				}
				if task.PackageType == "" {
					t.Error("filterServers() task missing PackageType")
				}
			}
		})
	}
}

func TestWatcher_poll(t *testing.T) {
	tests := []struct {
		name         string
		setupWatcher func(t *testing.T) (*Watcher, func())
		expectError  bool
		errorSubstr  string
	}{
		{
			name: "successful poll with no servers to generate",
			setupWatcher: func(t *testing.T) (*Watcher, func()) {
				cfg := createTestConfig(t)
				
				server := createMockServer(map[string]interface{}{
					"/v0/servers": v0.ServerListResponse{
						Servers: []v0.ServerJSON{},
						Metadata: &v0.Metadata{NextCursor: ""},
					},
				})
				
				client := createRegistryClient(t, server)
				watcher, _ := NewWatcher(cfg, client, generate.Options{DryRun: true})
				return watcher, server.Close
			},
			expectError: false,
		},
		{
			name: "registry fetch error",
			setupWatcher: func(t *testing.T) (*Watcher, func()) {
				cfg := createTestConfig(t)
				
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("registry connection failed"))
				}))
				
				client := createRegistryClient(t, server)
				watcher, _ := NewWatcher(cfg, client, generate.Options{DryRun: true})
				return watcher, server.Close
			},
			expectError: true,
			errorSubstr: "failed to fetch servers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, cleanup := tt.setupWatcher(t)
			defer cleanup()
			ctx := context.Background()

			err := watcher.poll(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("poll() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("poll() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("poll() unexpected error = %v", err)
			}
		})
	}
}

func TestWatcher_Run(t *testing.T) {
	tests := []struct {
		name        string
		setupTest   func(t *testing.T) (*Watcher, context.Context, context.CancelFunc, func())
		expectError bool
		errorSubstr string
	}{
		{
			name: "immediate context cancellation exits cleanly",
			setupTest: func(t *testing.T) (*Watcher, context.Context, context.CancelFunc, func()) {
				cfg := createTestConfig(t)
				
				server := createMockServer(map[string]interface{}{
					"/v0/servers": v0.ServerListResponse{
						Servers: []v0.ServerJSON{},
						Metadata: &v0.Metadata{NextCursor: ""},
					},
				})
				
				client := createRegistryClient(t, server)
				watcher, _ := NewWatcher(cfg, client, generate.Options{DryRun: true})
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return watcher, ctx, cancel, server.Close
			},
			expectError: true,
			errorSubstr: "graceful shutdown",
		},
		{
			name: "timeout after short duration",
			setupTest: func(t *testing.T) (*Watcher, context.Context, context.CancelFunc, func()) {
				cfg := createTestConfig(t)
				cfg.Watch.PollInterval = 30 // 30 second intervals
				
				server := createMockServer(map[string]interface{}{
					"/v0/servers": v0.ServerListResponse{
						Servers: []v0.ServerJSON{},
						Metadata: &v0.Metadata{NextCursor: ""},
					},
				})
				
				client := createRegistryClient(t, server)
				watcher, _ := NewWatcher(cfg, client, generate.Options{DryRun: true})
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				return watcher, ctx, cancel, server.Close
			},
			expectError: true,
			errorSubstr: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, ctx, cancel, cleanup := tt.setupTest(t)
			defer cleanup()
			defer cancel()

			err := watcher.Run(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Run() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Run() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("Run() unexpected error = %v", err)
			}
		})
	}
}

// Integration test that demonstrates the watcher working end-to-end
func TestWatcher_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := createTestConfig(t)
	cfg.Watch.PollInterval = 30 // Quick polling for test

	// Create mock server that returns test servers
	servers := []v0.ServerJSON{
		createTestServer("io.github.test", "server", "1.0.0", []model.Package{
			createTestPackage("npm", "test-server", "1.0.0"),
		}),
	}
	
	server := createMockServer(map[string]interface{}{
		"/v0/servers": v0.ServerListResponse{
			Servers: servers,
			Metadata: &v0.Metadata{NextCursor: ""},
		},
	})
	defer server.Close()
	
	client := createRegistryClient(t, server)

	watcher, err := NewWatcher(cfg, client, generate.Options{DryRun: true})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	// Run for a short time to test the loop
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = watcher.Run(ctx)
	if err == nil || !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Run() expected timeout error, got: %v", err)
	}

	// Verify that the poll happened by checking state
	lastPoll := watcher.state.GetLastPoll()
	if lastPoll.IsZero() {
		t.Error("Run() should have updated last poll time")
	}
}