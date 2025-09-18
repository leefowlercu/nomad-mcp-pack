package registry_watch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/leefowlercu/nomad-mcp-pack/tests/integration/common"
)

const (
	// MinPollInterval is the minimum poll interval the app allows
	MinPollInterval = 30 * time.Second
	// TestTimeout is the default timeout for watch tests
	TestTimeout = 2 * time.Minute
	// StateFileWait is how long to wait for state file operations
	StateFileWait = 45 * time.Second
	// FirstPollWait is how long to wait for the first poll cycle to complete
	FirstPollWait = 45 * time.Second
)

// WatchState represents the structure of the watch state file
type WatchState struct {
	LastPoll time.Time                `json:"last_poll"`
	Servers  map[string]*ServerState  `json:"servers"`
}

// ServerState represents a server entry in the state file
type ServerState struct {
	Namespace     string    `json:"namespace"`
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	PackageType   string    `json:"package_type"`
	TransportType string    `json:"transport_type"`
	UpdatedAt     time.Time `json:"updated_at"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// runWatchForDuration runs the watch command for a specific duration
func runWatchForDuration(ctx context.Context, duration time.Duration, args ...string) (*common.CommandResult, error) {
	// Build binary first
	binaryPath, err := common.GetBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to build binary: %w", err)
	}

	// Prepare command
	cmdArgs := append([]string{"watch"}, args...)
	cmd := exec.CommandContext(ctx, binaryPath, cmdArgs...)

	// Capture output
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	// Start command
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for timeout or completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var cmdErr error
	select {
	case <-timeoutCtx.Done():
		// Timeout reached, send signal to stop
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			cmd.Process.Kill()
		}
		<-done // Wait for process to actually exit
	case err := <-done:
		cmdErr = err
	}

	exitCode := 0
	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &common.CommandResult{
		Command:  binaryPath,
		Args:     cmdArgs,
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
		Error:    cmdErr,
	}, nil
}

// runWatchUntilGeneration runs watch until a specific number of servers are generated
func runWatchUntilGeneration(ctx context.Context, stateFile string, expectedCount int, args ...string) (*common.CommandResult, error) {
	// Build binary first
	binaryPath, err := common.GetBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to build binary: %w", err)
	}

	// Prepare command
	cmdArgs := append([]string{"watch"}, args...)
	cmd := exec.CommandContext(ctx, binaryPath, cmdArgs...)

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Monitor state file for expected count
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, TestTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	for {
		select {
		case <-timeoutCtx.Done():
			cmd.Process.Signal(syscall.SIGTERM)
			<-done
			return nil, fmt.Errorf("timeout waiting for %d servers to be generated", expectedCount)
		case err := <-done:
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}
			return &common.CommandResult{
				Command:  binaryPath,
				Args:     cmdArgs,
				ExitCode: exitCode,
				Error:    err,
			}, nil
		case <-ticker.C:
			if state, err := readStateFile(stateFile); err == nil {
				if len(state.Servers) >= expectedCount {
					// Found expected count, stop the command
					cmd.Process.Signal(syscall.SIGTERM)
					<-done
					return &common.CommandResult{
						Command:  binaryPath,
						Args:     cmdArgs,
						ExitCode: 0,
					}, nil
				}
			}
		}
	}
}

// readStateFile reads and parses the watch state file
func readStateFile(path string) (*WatchState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state WatchState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// countGeneratedPacks counts the number of pack directories or archives in the output directory
func countGeneratedPacks(outputDir string) (int, error) {
	entries, err := os.ReadDir(outputDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to read output directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it looks like a pack directory
			if strings.Contains(entry.Name(), "-") {
				count++
			}
		} else if strings.HasSuffix(entry.Name(), ".zip") {
			// Archive format
			count++
		}
	}

	return count, nil
}

// createPreExistingPacks creates pre-existing pack directories to test overwrite behavior
func createPreExistingPacks(outputDir string, serverNames []string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, serverName := range serverNames {
		// Convert server name to pack directory format
		packName := strings.ReplaceAll(serverName, "/", "-")
		packName = strings.ReplaceAll(packName, ".", "-")

		packDir := filepath.Join(outputDir, packName+"-test")
		if err := os.MkdirAll(packDir, 0755); err != nil {
			return fmt.Errorf("failed to create pack directory %s: %w", packDir, err)
		}

		// Create a dummy file
		dummyFile := filepath.Join(packDir, "dummy.txt")
		if err := os.WriteFile(dummyFile, []byte("existing pack"), 0644); err != nil {
			return fmt.Errorf("failed to create dummy file: %w", err)
		}
	}

	return nil
}

// extractLogsWithPattern extracts log lines matching a pattern from the logs
func extractLogsWithPattern(logs, pattern string) []string {
	regex := regexp.MustCompile(pattern)
	lines := strings.Split(logs, "\n")
	var matches []string

	for _, line := range lines {
		if regex.MatchString(line) {
			matches = append(matches, line)
		}
	}

	return matches
}

// waitForStateFile waits for the state file to be created or updated
func waitForStateFile(path string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("state file %s not created within %v", path, timeout)
}

// createCorruptedStateFile creates a state file with invalid JSON
func createCorruptedStateFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write invalid JSON
	invalidJSON := `{"last_poll": "invalid-time", "servers": {`
	return os.WriteFile(path, []byte(invalidJSON), 0644)
}

// getUniqueTestDir creates a unique directory for test isolation
func getUniqueTestDir(testName string) string {
	timestamp := time.Now().Unix()
	return filepath.Join(os.TempDir(), fmt.Sprintf("nomad-mcp-pack-test-%s-%d", testName, timestamp))
}

// shouldKeepTestOutput returns true if test output should be preserved
func shouldKeepTestOutput() bool {
	return os.Getenv("KEEP_WATCH_LOGS") == "true" || os.Getenv("KEEP_TEST_OUTPUT") == "true"
}

// getWatchTestTimeout returns the timeout for watch tests
func getWatchTestTimeout() time.Duration {
	if timeout := os.Getenv("WATCH_TEST_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			return d
		}
	}
	return TestTimeout
}