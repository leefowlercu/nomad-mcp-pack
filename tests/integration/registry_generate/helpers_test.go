package registry_generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leefowlercu/nomad-mcp-pack/tests/integration/common"
)

const (
	// TestTimeout is the default timeout for individual tests
	TestTimeout = 30 * time.Second
	// RegistryTimeout is the timeout for registry operations
	RegistryTimeout = 60 * time.Second
)

// CreateTestContext creates a context with timeout for tests
func CreateTestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = TestTimeout
	}
	return context.WithTimeout(context.Background(), timeout)
}

// CreateTestOutputDir creates a temporary directory for test output
func CreateTestOutputDir(t *testing.T, prefix string) string {
	t.Helper()

	outputDir, err := common.CreateTempDir(prefix + "-*")
	if err != nil {
		t.Fatalf("Failed to create test output directory: %v", err)
	}

	t.Cleanup(func() {
		if !common.ShouldKeepTestOutput() {
			if err := common.CleanupTempDir(outputDir); err != nil {
				t.Logf("Warning: failed to cleanup test directory %s: %v", outputDir, err)
			}
		} else {
			t.Logf("Keeping test output at: %s", outputDir)
		}
	})

	return outputDir
}

// RequireRegistryAvailable skips the test if registry is not available
func RequireRegistryAvailable(t *testing.T, registry *common.RegistryManager) {
	t.Helper()

	if !registry.IsAvailable() {
		t.Skip("Registry not available - skipping test")
	}
}

// VerifyGeneratedPackContent performs content validation on generated pack files
func VerifyGeneratedPackContent(t *testing.T, packDir string, serverName string, fixture TestServerFixture) {
	t.Helper()

	// Verify metadata.hcl contains expected content
	metadataPath := filepath.Join(packDir, "metadata.hcl")
	common.AssertFileContains(t, metadataPath, `name = "`+fixture.ExpectedPackName+`"`)
	common.AssertFileContains(t, metadataPath, `description = "Generated Nomad Pack for MCP Server: `+serverName+`"`)

	// Verify variables.hcl contains package-specific variables
	variablesPath := filepath.Join(packDir, "variables.hcl")
	switch fixture.PackageType {
	case "npm":
		common.AssertFileContains(t, variablesPath, `variable "npm_package"`)
		common.AssertFileContains(t, variablesPath, `default = "`+fixture.Identifier+`"`)
	case "pypi":
		common.AssertFileContains(t, variablesPath, `variable "pip_package"`)
		common.AssertFileContains(t, variablesPath, `default = "`+fixture.Identifier+`"`)
	}

	// Verify template contains package-specific job configuration
	templatePath := filepath.Join(packDir, "templates", "mcp-server.nomad.tpl")
	common.AssertFileExists(t, templatePath)

	// Verify README.md was generated
	readmePath := filepath.Join(packDir, "README.md")
	common.AssertFileContains(t, readmePath, serverName)
	common.AssertFileContains(t, readmePath, fixture.PackageType)
}

// VerifyArchiveContent validates the structure of a generated archive
func VerifyArchiveContent(t *testing.T, archivePath string) {
	t.Helper()

	// Verify the archive file exists and is not empty
	info, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("Archive file does not exist: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Archive file is empty")
	}

	if info.IsDir() {
		t.Error("Archive path is a directory, expected file")
	}

	// Note: Full archive content validation would require unzipping
	// which is beyond the scope of these integration tests
}

// ValidateCommandSuccess ensures command succeeded and provides detailed error info
func ValidateCommandSuccess(t *testing.T, result *common.CommandResult, operation string) {
	t.Helper()

	if !result.IsSuccess() {
		t.Errorf("%s failed:\nCommand: %s %v\nExit Code: %d\nStdout: %s\nStderr: %s\nError: %v\nDuration: %v",
			operation,
			result.Command,
			result.Args,
			result.ExitCode,
			result.Stdout,
			result.Stderr,
			result.Error,
			result.Duration,
		)
	}
}

// ValidateCommandFailure ensures command failed with expected behavior
func ValidateCommandFailure(t *testing.T, result *common.CommandResult, operation string, expectedError string) {
	t.Helper()

	if result.IsSuccess() {
		t.Errorf("%s should have failed but succeeded:\nCommand: %s %v\nStdout: %s\nStderr: %s",
			operation,
			result.Command,
			result.Args,
			result.Stdout,
			result.Stderr,
		)
		return
	}

	if expectedError != "" {
		output := result.GetOutput()
		if output == "" {
			t.Errorf("%s failed but produced no error output", operation)
		} else {
			common.AssertErrorContains(t, result, expectedError, operation)
		}
	}
}

// LogTestInfo logs useful information about the test environment
func LogTestInfo(t *testing.T, registryURL string, outputDir string) {
	t.Helper()

	t.Logf("Test environment:")
	t.Logf("  Registry URL: %s", registryURL)
	t.Logf("  Output directory: %s", outputDir)
	t.Logf("  Keep output: %v", common.ShouldKeepTestOutput())
}

// GetTestServerSpecs returns server specifications for testing
func GetTestServerSpecs() []string {
	return []string{
		"io.github.21st-dev/magic-mcp@0.0.1-seed",
		"io.github.adfin-engineering/mcp-server-adfin@0.0.1-seed",
		"io.github.21st-dev/magic-mcp@latest",
		"io.github.adfin-engineering/mcp-server-adfin@latest",
	}
}

// GetInvalidServerSpecs returns invalid server specifications for error testing
func GetInvalidServerSpecs() []string {
	return []string{
		"nonexistent-server@1.0.0",
		"invalid/format",
		"",
		"server-without-version",
		"@version-without-server",
	}
}