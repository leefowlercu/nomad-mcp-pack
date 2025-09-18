package common

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	// DefaultLocalRegistryURL is the default URL for local registry
	DefaultLocalRegistryURL = "http://localhost:8080/"
	// DefaultLiveRegistryURL is the default URL for live registry
	DefaultLiveRegistryURL = "https://registry.modelcontextprotocol.io/"
	// RegistryHealthTimeout is the timeout for registry health checks
	RegistryHealthTimeout = 60 * time.Second
	// RegistryStartupTimeout is the timeout for starting the registry
	RegistryStartupTimeout = 120 * time.Second
)

// RegistryManager manages local registry lifecycle for integration tests
type RegistryManager struct {
	URL         string
	isLocal     bool
	isStarted   bool
	registryDir string
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager() *RegistryManager {
	url := GetRegistryURL()
	isLocal := !IsLiveRegistry()

	// Find the registry directory (submodule)
	registryDir := findRegistryDir()

	return &RegistryManager{
		URL:         url,
		isLocal:     isLocal,
		registryDir: registryDir,
	}
}

// GetRegistryURL returns the registry URL to use for tests
func GetRegistryURL() string {
	if url := os.Getenv("INTEGRATION_REGISTRY_URL"); url != "" {
		return url
	}

	if IsLiveRegistry() {
		return DefaultLiveRegistryURL
	}

	return DefaultLocalRegistryURL
}

// IsLiveRegistry returns true if tests should use the live registry
func IsLiveRegistry() bool {
	return os.Getenv("USE_LIVE_REGISTRY") == "true"
}

// ShouldSkipRegistryTests returns true if registry tests should be skipped
func ShouldSkipRegistryTests() bool {
	return os.Getenv("SKIP_REGISTRY_TESTS") == "true"
}

// findRegistryDir finds the registry submodule directory
func findRegistryDir() string {
	// Try common locations
	candidates := []string{
		"tests/integration/registry",
		"./tests/integration/registry",
		"../registry",
	}

	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && stat.IsDir() {
			if absPath, err := filepath.Abs(candidate); err == nil {
				return absPath
			}
		}
	}

	return ""
}

// Start starts the local registry if needed
func (rm *RegistryManager) Start(ctx context.Context) error {
	if !rm.isLocal {
		log.Printf("Using live registry at %s", rm.URL)
		return rm.WaitForHealth(ctx)
	}

	if rm.isStarted {
		return nil
	}

	if rm.registryDir == "" {
		return fmt.Errorf("registry directory not found - ensure the MCP registry submodule is initialized")
	}

	log.Printf("Starting local registry from %s", rm.registryDir)

	// Check if Docker is available
	if !isDockerAvailable() {
		return fmt.Errorf("Docker is not available - required for local registry")
	}

	// Start the registry using docker-compose
	cmd := exec.CommandContext(ctx, "docker-compose", "up", "-d")
	cmd.Dir = rm.registryDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}

	rm.isStarted = true

	// Wait for the registry to be healthy
	return rm.WaitForHealth(ctx)
}

// Stop stops the local registry
func (rm *RegistryManager) Stop(ctx context.Context) error {
	if !rm.isLocal || !rm.isStarted {
		return nil
	}

	if rm.registryDir == "" {
		return nil
	}

	log.Printf("Stopping local registry")

	cmd := exec.CommandContext(ctx, "docker-compose", "down")
	cmd.Dir = rm.registryDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to stop registry: %v", err)
	}

	rm.isStarted = false
	return nil
}

// WaitForHealth waits for the registry to be healthy
func (rm *RegistryManager) WaitForHealth(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, RegistryHealthTimeout)
	defer cancel()

	log.Printf("Waiting for registry health at %s", rm.URL)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for registry health: %w", ctx.Err())
		case <-ticker.C:
			if rm.IsHealthy(client) {
				log.Printf("Registry is healthy at %s", rm.URL)
				return nil
			}
		}
	}
}

// IsHealthy checks if the registry is healthy
func (rm *RegistryManager) IsHealthy(client *http.Client) bool {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	// Try the health endpoint
	healthURL := fmt.Sprintf("%s/v0/health", rm.URL)
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	// Fallback: try listing servers
	serversURL := fmt.Sprintf("%s/v0/servers?limit=1", rm.URL)
	resp, err = client.Get(serversURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// IsAvailable checks if the registry is available without waiting
func (rm *RegistryManager) IsAvailable() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	return rm.IsHealthy(client)
}

// GetTestServers returns a list of known test servers available in the registry
func (rm *RegistryManager) GetTestServers(ctx context.Context) ([]TestServer, error) {
	// These are servers that should be available in the registry
	var knownServers []TestServer

	if rm.isLocal {
		// Local registry seed data
		knownServers = []TestServer{
			{
				Name:         "io.github.21st-dev/magic-mcp",
				Version:      "0.0.1-seed",
				PackageType:  "npm",
				Transport:    "stdio",
				Identifier:   "@21st-dev/magic",
			},
			{
				Name:         "io.github.adfin-engineering/mcp-server-adfin",
				Version:      "0.0.1-seed",
				PackageType:  "pypi",
				Transport:    "stdio",
				Identifier:   "adfinmcp",
			},
		}
	} else {
		// Live registry servers
		knownServers = []TestServer{
			{
				Name:         "io.github.kirbah/mcp-youtube",
				Version:      "0.2.6",
				PackageType:  "npm",
				Transport:    "stdio",
				Identifier:   "@kirbah/mcp-youtube",
			},
			{
				Name:         "io.github.huoshuiai42/huoshui-file-search",
				Version:      "1.0.0",
				PackageType:  "pypi",
				Transport:    "stdio",
				Identifier:   "huoshui-file-search",
			},
		}
	}

	// For now, return all known servers - let the tests handle availability
	// TODO: Implement proper server availability checking once API is clarified
	return knownServers, nil
}

// hasServer checks if a specific server exists in the registry
func (rm *RegistryManager) hasServer(ctx context.Context, name, version string) bool {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/v0/servers/%s?version=%s", rm.URL, name, version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// isDockerAvailable checks if Docker is available
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}

// TestServer represents a server available for testing
type TestServer struct {
	Name        string
	Version     string
	PackageType string
	Transport   string
	Identifier  string
}