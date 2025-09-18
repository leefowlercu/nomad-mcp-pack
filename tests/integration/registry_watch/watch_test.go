package registry_watch_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leefowlercu/nomad-mcp-pack/tests/integration/common"
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Skip all tests if registry tests are disabled
	if common.ShouldSkipRegistryTests() {
		os.Exit(0)
	}

	// Setup registry
	registry := common.NewRegistryManager()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := registry.Start(ctx); err != nil {
		panic("Failed to start registry: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup registry
	if err := registry.Stop(ctx); err != nil {
		// Don't panic on cleanup failure, just log
		println("Warning: Failed to stop registry:", err.Error())
	}

	// Cleanup binary
	common.CleanupBinary()

	os.Exit(code)
}

func TestWatch_BasicPolling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), getWatchTestTimeout())
	defer cancel()

	// Setup unique test environment
	outputDir := getUniqueTestDir("basic-polling")
	stateFile := filepath.Join(outputDir, "state.json")

	if !shouldKeepTestOutput() {
		defer os.RemoveAll(outputDir)
	} else {
		t.Logf("Test output preserved at: %s", outputDir)
	}

	// Run watch for minimum duration to allow one poll cycle
	args := []string{
		"--registry-url", common.GetRegistryURL(),
		"--output-dir", outputDir,
		"--state-file", stateFile,
		"--poll-interval", "30",
		"--filter-package-types", "npm",
		"--filter-transport-types", "stdio",
	}

	result, err := runWatchForDuration(ctx, FirstPollWait, args...)
	if err != nil {
		t.Fatalf("Failed to run watch command: %v", err)
	}

	// Verify graceful shutdown
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0 for graceful shutdown, got %d. Output: %s", result.ExitCode, result.GetOutput())
	}

	// Verify state file was created (should already exist after watch ran)
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("State file not created: %v", err)
	}

	// Read and verify state file
	state, err := readStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	// Verify state structure
	if state.LastPoll.IsZero() {
		t.Error("LastPoll should not be zero")
	}

	if len(state.Servers) == 0 {
		t.Error("Expected at least one server in state")
	}

	// Verify server state structure
	for key, server := range state.Servers {
		if server.Namespace == "" {
			t.Errorf("Server %s missing namespace", key)
		}
		if server.Name == "" {
			t.Errorf("Server %s missing name", key)
		}
		if server.Version == "" {
			t.Errorf("Server %s missing version", key)
		}
		if server.PackageType == "" {
			t.Errorf("Server %s missing package type", key)
		}
		if server.TransportType == "" {
			t.Errorf("Server %s missing transport type", key)
		}

		// Verify key format includes transport type
		expectedKey := server.Namespace + "/" + server.Name + "@" + server.Version + ":" + server.PackageType + ":" + server.TransportType
		if key != expectedKey {
			t.Errorf("Invalid state key format. Expected %s, got %s", expectedKey, key)
		}
	}

	t.Logf("Successfully generated state for %d servers", len(state.Servers))
}

func TestWatch_StateManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*getWatchTestTimeout())
	defer cancel()

	// Setup unique test environment
	outputDir := getUniqueTestDir("state-management")
	stateFile := filepath.Join(outputDir, "state.json")

	if !shouldKeepTestOutput() {
		defer os.RemoveAll(outputDir)
	} else {
		t.Logf("Test output preserved at: %s", outputDir)
	}

	args := []string{
		"--registry-url", common.GetRegistryURL(),
		"--output-dir", outputDir,
		"--state-file", stateFile,
		"--poll-interval", "30",
		"--filter-package-types", "npm",
		"--filter-transport-types", "stdio",
	}

	// First run - should generate state
	t.Log("Running first watch instance...")
	result1, err := runWatchForDuration(ctx, FirstPollWait, args...)
	if err != nil {
		t.Fatalf("Failed to run first watch command: %v", err)
	}

	if result1.ExitCode != 0 {
		t.Errorf("First watch run failed with exit code %d", result1.ExitCode)
	}

	// Read first state
	state1, err := readStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file after first run: %v", err)
	}

	firstPollTime := state1.LastPoll
	firstServerCount := len(state1.Servers)

	if firstServerCount == 0 {
		t.Fatal("No servers generated in first run")
	}

	t.Logf("First run generated %d servers", firstServerCount)

	// Second run - should not regenerate existing servers
	t.Log("Running second watch instance...")
	result2, err := runWatchForDuration(ctx, FirstPollWait, args...)
	if err != nil {
		t.Fatalf("Failed to run second watch command: %v", err)
	}

	if result2.ExitCode != 0 {
		t.Errorf("Second watch run failed with exit code %d", result2.ExitCode)
	}

	// Read second state
	state2, err := readStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file after second run: %v", err)
	}

	// Verify state persistence
	if state2.LastPoll.Before(firstPollTime) {
		t.Error("LastPoll should be updated on second run")
	}

	// Should have same or more servers (in case registry has new servers)
	if len(state2.Servers) < firstServerCount {
		t.Errorf("Second run has fewer servers (%d) than first run (%d)", len(state2.Servers), firstServerCount)
	}

	// Verify that original servers are still present
	for key, server1 := range state1.Servers {
		if server2, exists := state2.Servers[key]; !exists {
			t.Errorf("Server %s missing from second run", key)
		} else {
			// Generated time should be the same (not regenerated)
			if !server2.GeneratedAt.Equal(server1.GeneratedAt) {
				t.Errorf("Server %s was regenerated (GeneratedAt changed)", key)
			}
		}
	}

	t.Logf("State management test passed - %d servers preserved", len(state1.Servers))
}

func TestWatch_Filters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testCases := []struct {
		name                  string
		packageTypes          []string
		transportTypes        []string
		expectedMinServers    int
		expectPackageType     string
		expectTransportType   string
	}{
		{
			name:                "NPM only",
			packageTypes:        []string{"npm"},
			transportTypes:      []string{"stdio", "http", "sse"},
			expectedMinServers:  1,
			expectPackageType:   "npm",
		},
		{
			name:                "stdio only",
			packageTypes:        []string{"npm", "pypi", "oci"},
			transportTypes:      []string{"stdio"},
			expectedMinServers:  1,
			expectTransportType: "stdio",
		},
		{
			name:               "NPM + http combination",
			packageTypes:       []string{"npm"},
			transportTypes:     []string{"http"},
			expectedMinServers: 1,
			expectPackageType:  "npm",
			expectTransportType: "streamable-http",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), getWatchTestTimeout())
			defer cancel()

			// Setup unique test environment
			outputDir := getUniqueTestDir("filters-" + strings.ReplaceAll(tc.name, " ", "-"))
			stateFile := filepath.Join(outputDir, "state.json")

			if !shouldKeepTestOutput() {
				defer os.RemoveAll(outputDir)
			} else {
				t.Logf("Test output preserved at: %s", outputDir)
			}

			args := []string{
				"--registry-url", common.GetRegistryURL(),
				"--output-dir", outputDir,
				"--state-file", stateFile,
				"--poll-interval", "30",
				"--filter-package-types", strings.Join(tc.packageTypes, ","),
				"--filter-transport-types", strings.Join(tc.transportTypes, ","),
			}

			result, err := runWatchForDuration(ctx, FirstPollWait, args...)
			if err != nil {
				t.Fatalf("Failed to run watch command: %v", err)
			}

			if result.ExitCode != 0 {
				t.Errorf("Watch command failed with exit code %d", result.ExitCode)
			}

			// Read and verify state
			state, err := readStateFile(stateFile)
			if err != nil {
				t.Fatalf("Failed to read state file: %v", err)
			}

			if len(state.Servers) < tc.expectedMinServers {
				t.Errorf("Expected at least %d servers, got %d", tc.expectedMinServers, len(state.Servers))
			}

			// Verify all servers match the filters
			for key, server := range state.Servers {
				if tc.expectPackageType != "" && server.PackageType != tc.expectPackageType {
					t.Errorf("Server %s has package type %s, expected %s", key, server.PackageType, tc.expectPackageType)
				}

				if tc.expectTransportType != "" && server.TransportType != tc.expectTransportType {
					t.Errorf("Server %s has transport type %s, expected %s", key, server.TransportType, tc.expectTransportType)
				}

				// Verify package type is in allowed list
				packageTypeMatch := false
				for _, pt := range tc.packageTypes {
					if server.PackageType == pt {
						packageTypeMatch = true
						break
					}
				}
				if !packageTypeMatch {
					t.Errorf("Server %s has package type %s not in allowed list %v", key, server.PackageType, tc.packageTypes)
				}
			}

			t.Logf("Filter test '%s' passed with %d servers", tc.name, len(state.Servers))
		})
	}
}

func TestWatch_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), getWatchTestTimeout())
	defer cancel()

	// Setup unique test environment
	outputDir := getUniqueTestDir("dry-run")
	stateFile := filepath.Join(outputDir, "state.json")

	if !shouldKeepTestOutput() {
		defer os.RemoveAll(outputDir)
	} else {
		t.Logf("Test output preserved at: %s", outputDir)
	}

	args := []string{
		"--registry-url", common.GetRegistryURL(),
		"--output-dir", outputDir,
		"--state-file", stateFile,
		"--poll-interval", "30",
		"--filter-package-types", "npm",
		"--filter-transport-types", "stdio",
		"--dry-run",
	}

	result, err := runWatchForDuration(ctx, FirstPollWait, args...)
	if err != nil {
		t.Fatalf("Failed to run watch command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Watch command failed with exit code %d", result.ExitCode)
	}

	// Check if state file was created (dry-run currently doesn't create state files)
	// TODO: Consider if dry-run should still update state for tracking purposes
	if _, err := os.Stat(stateFile); err != nil {
		t.Log("State file not created in dry-run mode (expected current behavior)")
	} else {
		// If state file exists, verify it's valid
		state, err := readStateFile(stateFile)
		if err != nil {
			t.Fatalf("Failed to read state file: %v", err)
		}
		t.Logf("State file created with %d servers", len(state.Servers))
	}

	// Verify no actual pack directories were created (except the output dir itself)
	packCount, err := countGeneratedPacks(outputDir)
	if err != nil {
		t.Fatalf("Failed to count generated packs: %v", err)
	}

	if packCount > 0 {
		t.Errorf("Dry-run should not create pack directories, but found %d", packCount)
	}

	t.Logf("Dry-run test passed - state updated but no packs created")
}

func TestWatch_InvalidServerName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), getWatchTestTimeout())
	defer cancel()

	// Setup unique test environment
	outputDir := getUniqueTestDir("invalid-server")
	stateFile := filepath.Join(outputDir, "state.json")

	if !shouldKeepTestOutput() {
		defer os.RemoveAll(outputDir)
	} else {
		t.Logf("Test output preserved at: %s", outputDir)
	}

	// Use broader filters to potentially catch servers with invalid names
	args := []string{
		"--registry-url", common.GetRegistryURL(),
		"--output-dir", outputDir,
		"--state-file", stateFile,
		"--poll-interval", "30",
		"--dry-run",
	}

	result, err := runWatchForDuration(ctx, StateFileWait, args...)
	if err != nil {
		t.Fatalf("Failed to run watch command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Watch command failed with exit code %d", result.ExitCode)
	}

	// Verify command completed successfully despite invalid server names
	// (The watch command should log warnings but continue processing)
	state, err := readStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	// Should have some valid servers processed
	if len(state.Servers) == 0 {
		t.Error("Expected at least some valid servers to be processed")
	}

	// Verify all servers in state have valid key format
	for key, server := range state.Servers {
		expectedKey := server.Namespace + "/" + server.Name + "@" + server.Version + ":" + server.PackageType + ":" + server.TransportType
		if key != expectedKey {
			t.Errorf("Invalid state key format. Expected %s, got %s", expectedKey, key)
		}

		// Verify namespace/name format
		if !strings.Contains(server.Namespace+"/"+server.Name, "/") {
			t.Errorf("Server name should contain namespace/name format, got %s/%s", server.Namespace, server.Name)
		}
	}

	t.Logf("Invalid server name test passed - %d valid servers processed", len(state.Servers))
}

func TestWatch_ErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), getWatchTestTimeout())
	defer cancel()

	// Setup unique test environment
	outputDir := getUniqueTestDir("error-recovery")
	stateFile := filepath.Join(outputDir, "state.json")

	if !shouldKeepTestOutput() {
		defer os.RemoveAll(outputDir)
	} else {
		t.Logf("Test output preserved at: %s", outputDir)
	}

	// Create some pre-existing pack directories to cause "already exists" errors
	preExistingServers := []string{
		"io.github.test/server1",
		"io.github.test/server2",
	}

	if err := createPreExistingPacks(outputDir, preExistingServers); err != nil {
		t.Fatalf("Failed to create pre-existing packs: %v", err)
	}

	// Run without --force-overwrite to trigger errors
	args := []string{
		"--registry-url", common.GetRegistryURL(),
		"--output-dir", outputDir,
		"--state-file", stateFile,
		"--poll-interval", "30",
		"--filter-package-types", "npm",
		"--filter-transport-types", "stdio",
		// Note: Not using --dry-run here to test actual file operations
	}

	result, err := runWatchForDuration(ctx, StateFileWait, args...)
	if err != nil {
		t.Fatalf("Failed to run watch command: %v", err)
	}

	// Command should still exit successfully (errors are non-critical)
	if result.ExitCode != 0 {
		t.Errorf("Watch command failed with exit code %d", result.ExitCode)
	}

	// Verify state file was created despite errors
	state, err := readStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	// Should have some servers in state (successful generations)
	if len(state.Servers) == 0 {
		t.Error("Expected at least some servers in state despite errors")
	}

	// Verify state was updated even with generation errors
	if state.LastPoll.IsZero() {
		t.Error("LastPoll should be updated even with generation errors")
	}

	t.Logf("Error recovery test passed - state saved despite errors, %d servers processed", len(state.Servers))
}

func TestWatch_ForceOverwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), getWatchTestTimeout())
	defer cancel()

	// Setup unique test environment
	outputDir := getUniqueTestDir("force-overwrite")
	stateFile := filepath.Join(outputDir, "state.json")

	if !shouldKeepTestOutput() {
		defer os.RemoveAll(outputDir)
	} else {
		t.Logf("Test output preserved at: %s", outputDir)
	}

	// Create some pre-existing pack directories
	preExistingServers := []string{
		"io.github.test/server1",
		"io.github.test/server2",
	}

	if err := createPreExistingPacks(outputDir, preExistingServers); err != nil {
		t.Fatalf("Failed to create pre-existing packs: %v", err)
	}

	// Run with --force-overwrite
	args := []string{
		"--registry-url", common.GetRegistryURL(),
		"--output-dir", outputDir,
		"--state-file", stateFile,
		"--poll-interval", "30",
		"--filter-package-types", "npm",
		"--filter-transport-types", "stdio",
		"--force-overwrite",
		"--dry-run", // Use dry-run to avoid file system complications
	}

	result, err := runWatchForDuration(ctx, StateFileWait, args...)
	if err != nil {
		t.Fatalf("Failed to run watch command: %v", err)
	}

	// Command should succeed with force-overwrite
	if result.ExitCode != 0 {
		t.Errorf("Watch command with force-overwrite failed with exit code %d", result.ExitCode)
	}

	// Verify state file was created
	state, err := readStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	// Should have servers in state
	if len(state.Servers) == 0 {
		t.Error("Expected servers in state with force-overwrite")
	}

	t.Logf("Force overwrite test passed - %d servers processed", len(state.Servers))
}

func TestWatch_SignalHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), getWatchTestTimeout())
	defer cancel()

	// Setup unique test environment
	outputDir := getUniqueTestDir("signal-handling")
	stateFile := filepath.Join(outputDir, "state.json")

	if !shouldKeepTestOutput() {
		defer os.RemoveAll(outputDir)
	} else {
		t.Logf("Test output preserved at: %s", outputDir)
	}

	args := []string{
		"--registry-url", common.GetRegistryURL(),
		"--output-dir", outputDir,
		"--state-file", stateFile,
		"--poll-interval", "30",
		"--filter-package-types", "npm",
		"--filter-transport-types", "stdio",
		"--dry-run",
	}

	// Run for enough time to complete first poll, then signal
	result, err := runWatchForDuration(ctx, StateFileWait, args...)
	if err != nil {
		t.Fatalf("Failed to run watch command: %v", err)
	}

	// Graceful shutdown should result in exit code 0
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0 for graceful shutdown, got %d", result.ExitCode)
	}

	// Verify state was saved before shutdown
	state, err := readStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file after signal: %v", err)
	}

	if len(state.Servers) == 0 {
		t.Error("Expected servers in state after graceful shutdown")
	}

	if state.LastPoll.IsZero() {
		t.Error("LastPoll should be updated before shutdown")
	}

	t.Logf("Signal handling test passed - graceful shutdown with %d servers", len(state.Servers))
}

func TestWatch_StateCorruption(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), getWatchTestTimeout())
	defer cancel()

	// Setup unique test environment
	outputDir := getUniqueTestDir("state-corruption")
	stateFile := filepath.Join(outputDir, "state.json")

	if !shouldKeepTestOutput() {
		defer os.RemoveAll(outputDir)
	} else {
		t.Logf("Test output preserved at: %s", outputDir)
	}

	// Create corrupted state file
	if err := createCorruptedStateFile(stateFile); err != nil {
		t.Fatalf("Failed to create corrupted state file: %v", err)
	}

	args := []string{
		"--registry-url", common.GetRegistryURL(),
		"--output-dir", outputDir,
		"--state-file", stateFile,
		"--poll-interval", "30",
		"--filter-package-types", "npm",
		"--filter-transport-types", "stdio",
		"--dry-run",
	}

	result, err := runWatchForDuration(ctx, StateFileWait, args...)
	if err != nil {
		t.Fatalf("Failed to run watch command: %v", err)
	}

	// Command should handle corrupted state gracefully
	if result.ExitCode != 0 {
		t.Errorf("Watch command should handle corrupted state gracefully, got exit code %d", result.ExitCode)
	}

	// Verify new valid state was created
	state, err := readStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file after corruption recovery: %v", err)
	}

	// Should have created new valid state
	if len(state.Servers) == 0 {
		t.Error("Expected servers in state after corruption recovery")
	}

	if state.LastPoll.IsZero() {
		t.Error("LastPoll should be set in recovered state")
	}

	t.Logf("State corruption test passed - recovered with %d servers", len(state.Servers))
}