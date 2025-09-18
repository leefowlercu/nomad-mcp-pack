package common

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// CLITimeout is the default timeout for CLI commands
	CLITimeout = 60 * time.Second
	// BinaryName is the name of the nomad-mcp-pack binary
	BinaryName = "nomad-mcp-pack"
)

var (
	// binaryPath holds the path to the built binary
	binaryPath string
	// buildOnce ensures the binary is only built once
	buildOnce sync.Once
	// buildError holds any error from building the binary
	buildError error
)

// CommandResult represents the result of running a CLI command
type CommandResult struct {
	Command    string
	Args       []string
	ExitCode   int
	Stdout     string
	Stderr     string
	Duration   time.Duration
	Error      error
}

// IsSuccess returns true if the command succeeded
func (r *CommandResult) IsSuccess() bool {
	return r.ExitCode == 0 && r.Error == nil
}

// GetOutput returns the combined stdout and stderr
func (r *CommandResult) GetOutput() string {
	if r.Stdout != "" && r.Stderr != "" {
		return r.Stdout + "\n" + r.Stderr
	}
	if r.Stdout != "" {
		return r.Stdout
	}
	return r.Stderr
}

// CLIRunner provides utilities for running CLI commands
type CLIRunner struct {
	binaryPath string
	timeout    time.Duration
	env        []string
}

// NewCLIRunner creates a new CLI runner
func NewCLIRunner() (*CLIRunner, error) {
	path, err := GetBinaryPath()
	if err != nil {
		return nil, err
	}

	return &CLIRunner{
		binaryPath: path,
		timeout:    CLITimeout,
		env:        os.Environ(),
	}, nil
}

// WithTimeout sets a custom timeout for commands
func (r *CLIRunner) WithTimeout(timeout time.Duration) *CLIRunner {
	r.timeout = timeout
	return r
}

// WithEnv adds environment variables for commands
func (r *CLIRunner) WithEnv(key, value string) *CLIRunner {
	r.env = append(r.env, fmt.Sprintf("%s=%s", key, value))
	return r
}

// WithRegistryURL sets the registry URL environment variable
func (r *CLIRunner) WithRegistryURL(url string) *CLIRunner {
	return r.WithEnv("NOMAD_MCP_PACK_REGISTRY_URL", url)
}

// Run executes a command with the given arguments
func (r *CLIRunner) Run(ctx context.Context, args ...string) *CommandResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.binaryPath, args...)
	cmd.Env = r.env

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	result := &CommandResult{
		Command:  r.binaryPath,
		Args:     args,
		Duration: time.Since(start),
	}

	err := cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.Error = err

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	} else {
		result.ExitCode = 0
	}

	result.Duration = time.Since(start)
	return result
}

// RunGenerate runs the generate command with specific arguments
func (r *CLIRunner) RunGenerate(ctx context.Context, serverSpec string, opts ...GenerateOption) *CommandResult {
	args := []string{"generate", serverSpec}

	for _, opt := range opts {
		args = append(args, opt.Args()...)
	}

	return r.Run(ctx, args...)
}

// RunWatch runs the watch command with specific arguments
func (r *CLIRunner) RunWatch(ctx context.Context, opts ...WatchOption) *CommandResult {
	args := []string{"watch"}

	for _, opt := range opts {
		args = append(args, opt.Args()...)
	}

	return r.Run(ctx, args...)
}

// RunServer runs the server command with specific arguments
func (r *CLIRunner) RunServer(ctx context.Context, opts ...ServerOption) *CommandResult {
	args := []string{"server"}

	for _, opt := range opts {
		args = append(args, opt.Args()...)
	}

	return r.Run(ctx, args...)
}

// GenerateOption represents an option for the generate command
type GenerateOption interface {
	Args() []string
}

// WithOutputDir sets the output directory for generate command
type WithOutputDir string
func (w WithOutputDir) Args() []string { return []string{"--output-dir", string(w)} }

// WithOutputType sets the output type for generate command
type WithOutputType string
func (w WithOutputType) Args() []string { return []string{"--output-type", string(w)} }

// WithPackageType sets the package type for generate command
type WithPackageType string
func (w WithPackageType) Args() []string { return []string{"--package-type", string(w)} }

// WithTransportType sets the transport type for generate command
type WithTransportType string
func (w WithTransportType) Args() []string { return []string{"--transport-type", string(w)} }

// WithDryRun enables dry run mode
type WithDryRun struct{}
func (w WithDryRun) Args() []string { return []string{"--dry-run"} }

// WithForceOverwrite enables force overwrite mode
type WithForceOverwrite struct{}
func (w WithForceOverwrite) Args() []string { return []string{"--force-overwrite"} }

// WithAllowDeprecated allows deprecated servers
type WithAllowDeprecated struct{}
func (w WithAllowDeprecated) Args() []string { return []string{"--allow-deprecated"} }

// WatchOption represents an option for the watch command
type WatchOption interface {
	Args() []string
}

// WithWatchOutputDir sets the output directory for watch command
type WithWatchOutputDir string
func (w WithWatchOutputDir) Args() []string { return []string{"--output-dir", string(w)} }

// WithPollInterval sets the poll interval for watch command
type WithPollInterval int
func (w WithPollInterval) Args() []string { return []string{"--poll-interval", fmt.Sprintf("%d", int(w))} }

// WithFilterNames sets the filter names for watch command
type WithFilterNames []string
func (w WithFilterNames) Args() []string { return []string{"--filter-names", strings.Join(w, ",")} }

// ServerOption represents an option for the server command
type ServerOption interface {
	Args() []string
}

// WithServerAddr sets the server address
type WithServerAddr string
func (w WithServerAddr) Args() []string { return []string{"--addr", string(w)} }

// GetBinaryPath returns the path to the nomad-mcp-pack binary, building it if necessary
func GetBinaryPath() (string, error) {
	buildOnce.Do(func() {
		binaryPath, buildError = buildBinary()
	})
	return binaryPath, buildError
}

// buildBinary builds the nomad-mcp-pack binary for testing
func buildBinary() (string, error) {
	// Find the project root (where go.mod is)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", fmt.Errorf("failed to find project root: %w", err)
	}

	// Create a temporary directory for the binary
	tempDir, err := os.MkdirTemp("", "nomad-mcp-pack-test-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	binaryPath := filepath.Join(tempDir, BinaryName)

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./main.go")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to build binary: %w\nOutput: %s", err, output)
	}

	// Verify the binary was created
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf("binary not found after build: %w", err)
	}

	// Make the binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}

	return binaryPath, nil
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found")
}

// CleanupBinary removes the test binary
func CleanupBinary() {
	if binaryPath != "" {
		dir := filepath.Dir(binaryPath)
		os.RemoveAll(dir)
	}
}

// CreateTempDir creates a temporary directory for test output
func CreateTempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

// CleanupTempDir removes a temporary directory
func CleanupTempDir(dir string) error {
	if dir == "" {
		return nil
	}
	return os.RemoveAll(dir)
}

// ShouldKeepTestOutput returns true if test output should be kept for debugging
func ShouldKeepTestOutput() bool {
	return os.Getenv("KEEP_TEST_OUTPUT") == "true"
}