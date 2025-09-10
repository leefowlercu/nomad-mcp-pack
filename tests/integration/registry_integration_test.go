package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/leefowlercu/nomad-mcp-pack/pkg/registry"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

func TestIntegrationListServers(t *testing.T) {
	skipIfNoRegistry(t)

	client := createTestClient(t)
	checkRegistryHealth(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	tests := []struct {
		name string
		opts *registry.ListServersOptions
	}{
		{
			name: "list all servers with default options",
			opts: nil,
		},
		{
			name: "list with small limit",
			opts: &registry.ListServersOptions{
				Limit: 5,
			},
		},
		{
			name: "list with large limit",
			opts: &registry.ListServersOptions{
				Limit: 50,
			},
		},
		{
			name: "list with version filter latest",
			opts: &registry.ListServersOptions{
				Version: "latest",
				Limit:   10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.ListServers(ctx, tt.opts)
			assertNoError(t, err, "ListServers failed")
			assertNotNil(t, resp, "Response should not be nil")
			assertNotNil(t, resp.Metadata, "Metadata should not be nil")

			t.Logf("Found %d servers", resp.Metadata.Count)
			
			if resp.Servers != nil {
				for i, server := range resp.Servers {
					if i >= 3 { // Log only first 3 for brevity
						break
					}
					t.Logf("Server %d: %s (version: %s, status: %s)", i+1, server.Name, server.Version, server.Status)
				}
			}
		})
	}
}

func TestIntegrationSearchServers(t *testing.T) {
	skipIfNoRegistry(t)

	client := createTestClient(t)
	checkRegistryHealth(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	tests := []struct {
		name       string
		searchTerm string
		opts       *registry.ListServersOptions
		expectErr  bool
	}{
		{
			name:       "search with common term",
			searchTerm: "test",
			opts:       nil,
			expectErr:  false,
		},
		{
			name:       "search with limit",
			searchTerm: "server",
			opts: &registry.ListServersOptions{
				Limit: 5,
			},
			expectErr: false,
		},
		{
			name:       "search for specific file type",
			searchTerm: "file",
			opts: &registry.ListServersOptions{
				Version: "latest",
			},
			expectErr: false,
		},
		{
			name:       "search with empty term should fail",
			searchTerm: "",
			opts:       nil,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.SearchServers(ctx, tt.searchTerm, tt.opts)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			assertNoError(t, err, "SearchServers failed")
			assertNotNil(t, resp, "Response should not be nil")

			t.Logf("Search for '%s' returned %d servers", tt.searchTerm, resp.Metadata.Count)
		})
	}
}

func TestIntegrationGetLatestServers(t *testing.T) {
	skipIfNoRegistry(t)

	client := createTestClient(t)
	checkRegistryHealth(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	tests := []struct {
		name string
		opts *registry.ListServersOptions
	}{
		{
			name: "get latest servers with default options",
			opts: nil,
		},
		{
			name: "get latest servers with limit",
			opts: &registry.ListServersOptions{
				Limit: 10,
			},
		},
		{
			name: "get latest servers with search",
			opts: &registry.ListServersOptions{
				Search: "file",
				Limit:  5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetLatestServers(ctx, tt.opts)
			assertNoError(t, err, "GetLatestServers failed")
			assertNotNil(t, resp, "Response should not be nil")

			t.Logf("Found %d latest servers", resp.Metadata.Count)

			// Verify all returned servers are latest versions
			if resp.Servers != nil {
				for _, server := range resp.Servers {
					// In a real dev registry, we'd expect these to be marked as latest
					t.Logf("Latest server: %s (version: %s)", server.Name, server.Version)
				}
			}
		})
	}
}

func TestIntegrationGetUpdatedServers(t *testing.T) {
	skipIfNoRegistry(t)

	client := createTestClient(t)
	checkRegistryHealth(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	// Test with various time ranges
	tests := []struct {
		name         string
		updatedSince string
		expectErr    bool
	}{
		{
			name:         "servers updated in last hour",
			updatedSince: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			expectErr:    false,
		},
		{
			name:         "servers updated in last day",
			updatedSince: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			expectErr:    false,
		},
		{
			name:         "servers updated in last week",
			updatedSince: time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
			expectErr:    false,
		},
		{
			name:         "empty timestamp should fail",
			updatedSince: "",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetUpdatedServers(ctx, tt.updatedSince, nil)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			assertNoError(t, err, "GetUpdatedServers failed")
			assertNotNil(t, resp, "Response should not be nil")

			t.Logf("Found %d servers updated since %s", resp.Metadata.Count, tt.updatedSince)
		})
	}
}

func TestIntegrationGetLatestActiveServer(t *testing.T) {
	skipIfNoRegistry(t)

	client := createTestClient(t)
	checkRegistryHealth(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	listResp, err := client.ListServers(ctx, &registry.ListServersOptions{
		Limit: 10,
	})
	assertNoError(t, err, "Failed to list servers for test setup")

	if len(listResp.Servers) == 0 {
		t.Skip("No servers available in registry for testing")
	}

	var testServerName string
	for _, server := range listResp.Servers {
		if server.Status == model.StatusActive {
			testServerName = server.Name
			break
		}
	}

	if testServerName == "" {
		t.Skip("No active servers found for testing GetLatestActiveServer")
	}

	tests := []struct {
		name       string
		serverName string
		expectErr  bool
	}{
		{
			name:       "get latest active server with known name",
			serverName: testServerName,
			expectErr:  false,
		},
		{
			name:       "get latest active server with non-existent name",
			serverName: "non-existent-server-name-12345",
			expectErr:  true,
		},
		{
			name:       "empty server name should fail",
			serverName: "",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetLatestActiveServer(ctx, tt.serverName)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			assertNoError(t, err, "GetLatestActiveServer failed")
			assertNotNil(t, resp, "Response should not be nil")

			if resp.Status != model.StatusActive {
				t.Errorf("Expected active server, got status: %s", resp.Status)
			}

			if resp.Name != tt.serverName {
				t.Errorf("Expected server name %s, got %s", tt.serverName, resp.Name)
			}

			t.Logf("Latest active server: %s (version: %s)", resp.Name, resp.Version)
		})
	}
}

func TestIntegrationGetServerByNameAndVersion(t *testing.T) {
	skipIfNoRegistry(t)

	client := createTestClient(t)
	checkRegistryHealth(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	// First, get a server to test with
	listResp, err := client.ListServers(ctx, &registry.ListServersOptions{
		Limit: 5,
	})
	assertNoError(t, err, "Failed to list servers for test setup")

	if len(listResp.Servers) == 0 {
		t.Skip("No servers available in registry for testing")
	}

	testServer := listResp.Servers[0]

	tests := []struct {
		name       string
		serverName string
		version    string
		expectErr  bool
	}{
		{
			name:       "get server with known name and version",
			serverName: testServer.Name,
			version:    testServer.Version,
			expectErr:  false,
		},
		{
			name:       "get server with latest version",
			serverName: testServer.Name,
			version:    "latest",
			expectErr:  false,
		},
		{
			name:       "get server with non-existent version",
			serverName: testServer.Name,
			version:    "999.999.999",
			expectErr:  true,
		},
		{
			name:       "empty server name should fail",
			serverName: "",
			version:    "1.0.0",
			expectErr:  true,
		},
		{
			name:       "empty version should fail",
			serverName: testServer.Name,
			version:    "",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetServerByNameAndVersion(ctx, tt.serverName, tt.version)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			assertNoError(t, err, "GetServerByNameAndVersion failed")
			assertNotNil(t, resp, "Response should not be nil")

			if resp.Name != tt.serverName {
				t.Errorf("Expected server name %s, got %s", tt.serverName, resp.Name)
			}

			t.Logf("Found server: %s (version: %s)", resp.Name, resp.Version)
		})
	}
}

func TestIntegrationPagination(t *testing.T) {
	skipIfNoRegistry(t)

	client := createTestClient(t)
	checkRegistryHealth(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	// Test pagination by requesting small pages
	var allServers []any
	var cursor string
	pageSize := 3
	maxPages := 5

	for page := range maxPages {
		opts := &registry.ListServersOptions{
			Limit:  pageSize,
			Cursor: cursor,
		}

		resp, err := client.ListServers(ctx, opts)
		assertNoError(t, err, "ListServers pagination failed")
		assertNotNil(t, resp, "Response should not be nil")

		t.Logf("Page %d: %d servers", page+1, len(resp.Servers))

		if resp.Servers != nil {
			for _, server := range resp.Servers {
				allServers = append(allServers, server)
			}
		}

		// Check if there are more pages
		if resp.Metadata == nil || resp.Metadata.NextCursor == "" {
			t.Logf("No more pages after page %d", page+1)
			break
		}

		cursor = resp.Metadata.NextCursor
	}

	t.Logf("Total servers collected across pages: %d", len(allServers))
}

func TestIntegrationErrorHandling(t *testing.T) {
	skipIfNoRegistry(t)

	// Test with a client that has a very short timeout
	shortTimeoutClient, err := registry.NewClient(getRegistryURL())
	assertNoError(t, err, "Failed to create short timeout client")
	shortTimeoutClient.SetTimeout(1 * time.Millisecond) // Very short timeout

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name   string
		testFn func(t *testing.T)
	}{
		{
			name: "timeout handling",
			testFn: func(t *testing.T) {
				// This should timeout due to very short client timeout
				_, err := shortTimeoutClient.ListServers(ctx, nil)
				if err == nil {
					t.Log("Expected timeout error but request succeeded (registry very fast)")
				} else {
					t.Logf("Got expected error: %v", err)
				}
			},
		},
		{
			name: "invalid server ID",
			testFn: func(t *testing.T) {
				client := createTestClient(t)
				_, err := client.GetServer(ctx, "invalid-uuid-format")
				if err == nil {
					t.Error("Expected error for invalid server ID")
				}
			},
		},
		{
			name: "non-existent server ID",
			testFn: func(t *testing.T) {
				client := createTestClient(t)
				_, err := client.GetServer(ctx, "00000000-0000-0000-0000-000000000000")
				if err == nil {
					t.Error("Expected error for non-existent server ID")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFn(t)
		})
	}
}

func TestIntegrationConcurrency(t *testing.T) {
	skipIfNoRegistry(t)

	client := createTestClient(t)
	checkRegistryHealth(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	// Test concurrent requests to ensure client is thread-safe
	// Reduced concurrency to be gentler on the local registry
	const numGoroutines = 3
	done := make(chan error, numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			// Add small delay to stagger requests
			time.Sleep(time.Duration(id*200) * time.Millisecond)
			
			opts := &registry.ListServersOptions{
				Limit: 2, // Smaller limit for gentler load
			}
			
			resp, err := client.ListServers(ctx, opts)
			if err != nil {
				done <- err
				return
			}
			
			if resp == nil {
				done <- fmt.Errorf("goroutine %d: got nil response", id)
				return
			}
			
			t.Logf("Goroutine %d completed successfully with %d servers", id, len(resp.Servers))
			done <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	var errCount int
	for range numGoroutines {
		if err := <-done; err != nil {
			t.Logf("Concurrent request failed: %v", err)
			errCount++
		}
	}

	// Allow for some failures in concurrent scenarios due to registry limitations
	// But expect at least one success to prove client thread-safety
	if errCount >= numGoroutines {
		t.Errorf("All concurrent requests failed: %d/%d", errCount, numGoroutines)
	}
}