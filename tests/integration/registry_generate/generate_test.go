package registry_generate_test

import (
	"context"
	"os"
	"path/filepath"
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

func TestGenerateFromRegistry_BasicNPM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	runner, err := common.NewCLIRunner()
	if err != nil {
		t.Fatalf("Failed to create CLI runner: %v", err)
	}

	registry := common.NewRegistryManager()
	runner = runner.WithRegistryURL(registry.URL)

	// Create temp output directory
	outputDir, err := common.CreateTempDir("nomad-mcp-pack-test-npm-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if !common.ShouldKeepTestOutput() {
			common.CleanupTempDir(outputDir)
		} else {
			t.Logf("Test output kept at: %s", outputDir)
		}
	}()

	// Get test servers
	testServers, err := registry.GetTestServers(ctx)
	if err != nil {
		t.Fatalf("Failed to get test servers: %v", err)
	}

	// Find an NPM server
	var npmServer *common.TestServer
	for _, server := range testServers {
		if server.PackageType == "npm" {
			npmServer = &server
			break
		}
	}

	if npmServer == nil {
		t.Skip("No NPM servers available for testing")
	}

	serverSpec := npmServer.Name + "@" + npmServer.Version

	// Run generate command
	result := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("packdir"),
		common.WithPackageType(npmServer.PackageType),
		common.WithTransportType(npmServer.Transport),
		common.WithForceOverwrite{},
	)

	// Assert command succeeded
	common.AssertCommandSuccess(t, result, "Generate command should succeed for NPM package")

	// Compute expected pack name
	packName := common.ComputePackName(npmServer.Name, npmServer.Version, npmServer.PackageType, npmServer.Transport)

	// Assert pack was generated
	common.AssertPackGenerated(t, outputDir, packName, "NPM pack should be generated")

	// Assert pack structure
	packDir := filepath.Join(outputDir, packName)
	common.AssertPackStructure(t, packDir, "NPM pack should have correct structure")

	// Assert specific content for NPM packages
	jobTemplatePath := filepath.Join(packDir, "templates", "mcp-server.nomad.tpl")
	common.AssertFileContains(t, jobTemplatePath, npmServer.Identifier, "Job template should contain NPM package identifier")
	common.AssertFileContains(t, jobTemplatePath, "npm install", "Job template should contain npm install command")

	// Assert metadata contains server information
	metadataPath := filepath.Join(packDir, "metadata.hcl")
	common.AssertFileContains(t, metadataPath, npmServer.Version, "Metadata should contain server version")
	common.AssertFileContains(t, metadataPath, "pack {", "Metadata should have pack block")
}

func TestGenerateFromRegistry_BasicPyPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	runner, err := common.NewCLIRunner()
	if err != nil {
		t.Fatalf("Failed to create CLI runner: %v", err)
	}

	registry := common.NewRegistryManager()
	runner = runner.WithRegistryURL(registry.URL)

	// Create temp output directory
	outputDir, err := common.CreateTempDir("nomad-mcp-pack-test-pypi-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if !common.ShouldKeepTestOutput() {
			common.CleanupTempDir(outputDir)
		} else {
			t.Logf("Test output kept at: %s", outputDir)
		}
	}()

	// Get test servers
	testServers, err := registry.GetTestServers(ctx)
	if err != nil {
		t.Fatalf("Failed to get test servers: %v", err)
	}

	// Find a PyPI server
	var pypiServer *common.TestServer
	for _, server := range testServers {
		if server.PackageType == "pypi" {
			pypiServer = &server
			break
		}
	}

	if pypiServer == nil {
		t.Skip("No PyPI servers available for testing")
	}

	serverSpec := pypiServer.Name + "@" + pypiServer.Version

	// Run generate command
	result := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("packdir"),
		common.WithPackageType(pypiServer.PackageType),
		common.WithTransportType(pypiServer.Transport),
		common.WithForceOverwrite{},
	)

	// Assert command succeeded
	common.AssertCommandSuccess(t, result, "Generate command should succeed for PyPI package")

	// Compute expected pack name
	packName := common.ComputePackName(pypiServer.Name, pypiServer.Version, pypiServer.PackageType, pypiServer.Transport)

	// Assert pack was generated
	common.AssertPackGenerated(t, outputDir, packName, "PyPI pack should be generated")

	// Assert pack structure
	packDir := filepath.Join(outputDir, packName)
	common.AssertPackStructure(t, packDir, "PyPI pack should have correct structure")

	// Assert specific content for PyPI packages
	jobTemplatePath := filepath.Join(packDir, "templates", "mcp-server.nomad.tpl")
	common.AssertFileContains(t, jobTemplatePath, pypiServer.Identifier, "Job template should contain PyPI package identifier")
	common.AssertFileContains(t, jobTemplatePath, "pip install", "Job template should contain pip install command")

	// Assert metadata contains server information
	metadataPath := filepath.Join(packDir, "metadata.hcl")
	common.AssertFileContains(t, metadataPath, pypiServer.Version, "Metadata should contain server version")
	common.AssertFileContains(t, metadataPath, "pack {", "Metadata should have pack block")
}

func TestGenerateFromRegistry_LatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	runner, err := common.NewCLIRunner()
	if err != nil {
		t.Fatalf("Failed to create CLI runner: %v", err)
	}

	registry := common.NewRegistryManager()
	runner = runner.WithRegistryURL(registry.URL)

	// Create temp output directory
	outputDir, err := common.CreateTempDir("nomad-mcp-pack-test-latest-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if !common.ShouldKeepTestOutput() {
			common.CleanupTempDir(outputDir)
		} else {
			t.Logf("Test output kept at: %s", outputDir)
		}
	}()

	// Get test servers
	testServers, err := registry.GetTestServers(ctx)
	if err != nil {
		t.Fatalf("Failed to get test servers: %v", err)
	}

	if len(testServers) == 0 {
		t.Skip("No test servers available")
	}

	// Use the first available server with @latest
	testServer := &testServers[0]
	serverSpec := testServer.Name + "@latest"

	// Run generate command
	result := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("packdir"),
		common.WithPackageType(testServer.PackageType),
		common.WithTransportType(testServer.Transport),
		common.WithForceOverwrite{},
	)

	// Assert command succeeded
	common.AssertCommandSuccess(t, result, "Generate command should succeed with @latest version")

	// Note: We can't predict the exact pack name since @latest resolves to the actual latest version
	// So we just verify that some pack was generated
	common.AssertDirNotEmpty(t, outputDir, "Output directory should contain generated pack")
}

func TestGenerateFromRegistry_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	runner, err := common.NewCLIRunner()
	if err != nil {
		t.Fatalf("Failed to create CLI runner: %v", err)
	}

	registry := common.NewRegistryManager()
	runner = runner.WithRegistryURL(registry.URL)

	// Create temp output directory
	outputDir, err := common.CreateTempDir("nomad-mcp-pack-test-dryrun-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if !common.ShouldKeepTestOutput() {
			common.CleanupTempDir(outputDir)
		} else {
			t.Logf("Test output kept at: %s", outputDir)
		}
	}()

	// Get test servers
	testServers, err := registry.GetTestServers(ctx)
	if err != nil {
		t.Fatalf("Failed to get test servers: %v", err)
	}

	if len(testServers) == 0 {
		t.Skip("No test servers available")
	}

	testServer := &testServers[0]
	serverSpec := testServer.Name + "@" + testServer.Version

	// Run generate command with dry run
	result := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("packdir"),
		common.WithPackageType(testServer.PackageType),
		common.WithTransportType(testServer.Transport),
		common.WithDryRun{},
	)

	// Assert command succeeded
	common.AssertCommandSuccess(t, result, "Generate command should succeed in dry run mode")

	// Assert output mentions what would be created
	common.AssertOutputContains(t, result, "Would create pack", "Dry run should show what would be created")

	// Assert no files were actually created
	common.AssertDirEmpty(t, outputDir, "Dry run should not create any files")
}

func TestGenerateFromRegistry_Archive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	runner, err := common.NewCLIRunner()
	if err != nil {
		t.Fatalf("Failed to create CLI runner: %v", err)
	}

	registry := common.NewRegistryManager()
	runner = runner.WithRegistryURL(registry.URL)

	// Create temp output directory
	outputDir, err := common.CreateTempDir("nomad-mcp-pack-test-archive-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if !common.ShouldKeepTestOutput() {
			common.CleanupTempDir(outputDir)
		} else {
			t.Logf("Test output kept at: %s", outputDir)
		}
	}()

	// Get test servers
	testServers, err := registry.GetTestServers(ctx)
	if err != nil {
		t.Fatalf("Failed to get test servers: %v", err)
	}

	if len(testServers) == 0 {
		t.Skip("No test servers available")
	}

	testServer := &testServers[0]
	serverSpec := testServer.Name + "@" + testServer.Version

	// Run generate command with archive output
	result := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("archive"),
		common.WithPackageType(testServer.PackageType),
		common.WithTransportType(testServer.Transport),
		common.WithForceOverwrite{},
	)

	// Assert command succeeded
	common.AssertCommandSuccess(t, result, "Generate command should succeed with archive output")

	// Compute expected pack name
	packName := common.ComputePackName(testServer.Name, testServer.Version, testServer.PackageType, testServer.Transport)

	// Assert archive was generated
	common.AssertArchiveGenerated(t, outputDir, packName, "Archive should be generated")
}

func TestGenerateFromRegistry_InvalidServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runner, err := common.NewCLIRunner()
	if err != nil {
		t.Fatalf("Failed to create CLI runner: %v", err)
	}

	registry := common.NewRegistryManager()
	runner = runner.WithRegistryURL(registry.URL)

	// Create temp output directory
	outputDir, err := common.CreateTempDir("nomad-mcp-pack-test-invalid-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if !common.ShouldKeepTestOutput() {
			common.CleanupTempDir(outputDir)
		}
	}()

	// Use a non-existent server
	serverSpec := "non-existent/server@1.0.0"

	// Run generate command
	result := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("packdir"),
	)

	// Assert command failed
	common.AssertCommandFailure(t, result, "Generate command should fail for non-existent server")

	// Assert error message is helpful
	common.AssertErrorContains(t, result, "no server", "Error should mention server not found")
}

func TestGenerateFromRegistry_InvalidFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runner, err := common.NewCLIRunner()
	if err != nil {
		t.Fatalf("Failed to create CLI runner: %v", err)
	}

	registry := common.NewRegistryManager()
	runner = runner.WithRegistryURL(registry.URL)

	tests := []struct {
		name       string
		serverSpec string
		expectError string
	}{
		{
			name:       "missing version",
			serverSpec: "some-server",
			expectError: "exactly one '@' separator",
		},
		{
			name:       "empty server name",
			serverSpec: "@1.0.0",
			expectError: "server name stanza must not be empty",
		},
		{
			name:       "empty version",
			serverSpec: "some/server@",
			expectError: "server version stanza must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp output directory
			outputDir, err := common.CreateTempDir("nomad-mcp-pack-test-invalid-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer common.CleanupTempDir(outputDir)

			// Run generate command
			result := runner.RunGenerate(ctx, tt.serverSpec,
				common.WithOutputDir(outputDir),
				common.WithOutputType("packdir"),
			)

			// Assert command failed
			common.AssertCommandFailure(t, result, "Generate command should fail for invalid server spec")

			// Assert error message contains expected text
			common.AssertErrorContains(t, result, tt.expectError, "Error should contain expected message")
		})
	}
}

func TestGenerateFromRegistry_ForceOverwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	runner, err := common.NewCLIRunner()
	if err != nil {
		t.Fatalf("Failed to create CLI runner: %v", err)
	}

	registry := common.NewRegistryManager()
	runner = runner.WithRegistryURL(registry.URL)

	// Create temp output directory
	outputDir, err := common.CreateTempDir("nomad-mcp-pack-test-overwrite-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if !common.ShouldKeepTestOutput() {
			common.CleanupTempDir(outputDir)
		} else {
			t.Logf("Test output kept at: %s", outputDir)
		}
	}()

	// Get test servers
	testServers, err := registry.GetTestServers(ctx)
	if err != nil {
		t.Fatalf("Failed to get test servers: %v", err)
	}

	if len(testServers) == 0 {
		t.Skip("No test servers available")
	}

	testServer := &testServers[0]
	serverSpec := testServer.Name + "@" + testServer.Version

	// First generation (should succeed)
	result1 := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("packdir"),
		common.WithPackageType(testServer.PackageType),
		common.WithTransportType(testServer.Transport),
	)

	common.AssertCommandSuccess(t, result1, "First generate command should succeed")

	// Second generation without force (should fail)
	result2 := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("packdir"),
		common.WithPackageType(testServer.PackageType),
		common.WithTransportType(testServer.Transport),
	)

	common.AssertCommandFailure(t, result2, "Second generate without force should fail")
	common.AssertErrorContains(t, result2, "already exists", "Error should mention pack already exists")

	// Third generation with force (should succeed)
	result3 := runner.RunGenerate(ctx, serverSpec,
		common.WithOutputDir(outputDir),
		common.WithOutputType("packdir"),
		common.WithPackageType(testServer.PackageType),
		common.WithTransportType(testServer.Transport),
		common.WithForceOverwrite{},
	)

	common.AssertCommandSuccess(t, result3, "Generate with force overwrite should succeed")
}