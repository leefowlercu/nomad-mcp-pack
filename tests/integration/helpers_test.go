package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/leefowlercu/nomad-mcp-pack/pkg/registry"
)

const (
	defaultRegistryURL = "http://localhost:8080"
	defaultTimeout     = 30 * time.Second
	healthCheckRetries = 10
	healthCheckDelay   = 2 * time.Second
)

func skipIfNoRegistry(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Integration tests disabled via SKIP_INTEGRATION_TESTS")
	}

	if !isRegistryAvailable() {
		t.Skip("Local registry not available at " + getRegistryURL())
	}
}

func getRegistryURL() string {
	if url := os.Getenv("INTEGRATION_TEST_REGISTRY_URL"); url != "" {
		return url
	}
	return defaultRegistryURL
}

func getTestTimeout() time.Duration {
	if timeout := os.Getenv("INTEGRATION_TEST_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			return d
		}
	}
	return defaultTimeout
}

func createTestClient(t *testing.T) *registry.Client {
	t.Helper()

	client, err := registry.NewClient(getRegistryURL())
	if err != nil {
		t.Fatalf("Failed to create registry client: %v", err)
	}

	client.SetTimeout(getTestTimeout())
	return client
}

func isRegistryAvailable() bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("%s/v0/health", getRegistryURL())
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func waitForRegistry(t *testing.T) {
	t.Helper()

	for i := 0; i < healthCheckRetries; i++ {
		if isRegistryAvailable() {
			return
		}
		if i < healthCheckRetries-1 {
			t.Logf("Registry not ready, retrying in %v... (%d/%d)", healthCheckDelay, i+1, healthCheckRetries)
			time.Sleep(healthCheckDelay)
		}
	}

	t.Fatalf("Registry did not become available after %d retries", healthCheckRetries)
}

func checkRegistryHealth(t *testing.T, client *registry.Client) {
	t.Helper()

	// Test that we can reach the registry by attempting a simple list operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := &registry.ListServersOptions{
		Limit: 1, // Minimal request
	}

	_, err := client.ListServers(ctx, opts)
	if err != nil {
		t.Fatalf("Registry health check failed: %v", err)
	}
}

func cleanupTestData(t *testing.T) {
	t.Helper()
	// For read-only tests against dev registry, no cleanup needed
	// This function is here for future expansion if tests start creating data
}

func assertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func assertNotNil(t *testing.T, value any, msg string) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s: expected non-nil value", msg)
	}
}

func assertEmpty(t *testing.T, slice any, msg string) {
	t.Helper()
	switch s := slice.(type) {
	case []any:
		if len(s) > 0 {
			t.Fatalf("%s: expected empty slice, got %d items", msg, len(s))
		}
	default:
		t.Fatalf("%s: unsupported type for assertEmpty", msg)
	}
}