package generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/leefowlercu/nomad-mcp-pack/internal/serversearchutils"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

// Test fixtures and mock data

func mockServerJSON() *v0.ServerJSON {
	return &v0.ServerJSON{
		Name:        "test-server",
		Description: "Test MCP Server for unit testing",
		Version:     "1.0.0",
		Repository: model.Repository{
			URL:    "https://github.com/test/repo",
			Source: "github",
		},
		Packages: []model.Package{
			{
				RegistryType: "npm",
				Identifier:   "@test/package",
				Version:      "1.0.0",
				EnvironmentVariables: []model.KeyValueInput{
					{
						Name: "API_KEY",
						InputWithVariables: model.InputWithVariables{
							Input: model.Input{
								Description: "API Key for authentication",
								IsSecret:    true,
								IsRequired:  true,
							},
						},
					},
					{
						Name: "DEBUG",
						InputWithVariables: model.InputWithVariables{
							Input: model.Input{
								Description: "Enable debug mode",
								Default:     "false",
							},
						},
					},
				},
				RuntimeArguments: []model.Argument{
					{
						Type: model.ArgumentTypeNamed,
						Name: "port",
						InputWithVariables: model.InputWithVariables{
							Input: model.Input{
								Description: "Server port",
								Default:     "3000",
							},
						},
					},
				},
			},
			{
				RegistryType: "oci",
				Identifier:   "test/image",
				Version:      "latest",
			},
		},
	}
}

func mockServerSpec() *serversearchutils.ServerSearchSpec {
	return &serversearchutils.ServerSearchSpec{
		Namespace: "test.server",
		Name:      "name",
		Version:    "1.0.0",
	}
}

func mockMinimalServerJSON() *v0.ServerJSON {
	return &v0.ServerJSON{
		Name:        "minimal-server",
		Description: "Minimal test server",
		Version:     "0.1.0",
		Packages: []model.Package{
			{
				RegistryType: "npm",
				Identifier:   "minimal-package",
				Version:      "0.1.0",
			},
		},
	}
}

// Test helper functions

func setupTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "nomad-mcp-pack-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func assertFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("File does not exist: %s", path)
	}
}

func assertFileContains(t *testing.T, path, expected string) {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	if !contains(string(content), expected) {
		t.Errorf("File %s does not contain expected content %q", path, expected)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for Run function

func TestRun_NilServer(t *testing.T) {
	ctx := context.Background()
	serverSpec := mockServerSpec()
	opts := Options{
		OutputDir:  "./test",
		OutputType: "packdir",
	}

	err := Run(ctx, nil, serverSpec, "npm", opts)
	if err == nil {
		t.Error("Expected error for nil server, got none")
	}
	if !contains(err.Error(), "server cannot be nil") {
		t.Errorf("Expected error message about nil server, got: %v", err)
	}
}

func TestRun_NilServerSpec(t *testing.T) {
	ctx := context.Background()
	server := mockServerJSON()
	opts := Options{
		OutputDir:  "./test",
		OutputType: "packdir",
	}

	err := Run(ctx, server, nil, "npm", opts)
	if err == nil {
		t.Error("Expected error for nil serverSpec, got none")
	}
	if !contains(err.Error(), "serverSpec cannot be nil") {
		t.Errorf("Expected error message about nil serverSpec, got: %v", err)
	}
}

func TestRun_NoMatchingPackage(t *testing.T) {
	ctx := context.Background()
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	opts := Options{
		OutputDir:  "./test",
		OutputType: "packdir",
	}

	err := Run(ctx, server, serverSpec, "pypi", opts)
	if err == nil {
		t.Error("Expected error for non-existent package type, got none")
	}
	if !contains(err.Error(), "no package found for type pypi") {
		t.Errorf("Expected error message about missing package type, got: %v", err)
	}
}

func TestRun_Success_PackdirDryRun(t *testing.T) {
	ctx := context.Background()
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	opts := Options{
		OutputDir:  tempDir,
		OutputType: "packdir",
		DryRun:     true,
	}

	err := Run(ctx, server, serverSpec, "npm", opts)
	if err != nil {
		t.Errorf("Unexpected error for dry run: %v", err)
	}
}

func TestRun_Success_PackdirGeneration(t *testing.T) {
	ctx := context.Background()
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	opts := Options{
		OutputDir:  tempDir,
		OutputType: "packdir",
		DryRun:     false,
	}

	err := Run(ctx, server, serverSpec, "npm", opts)
	if err != nil {
		t.Errorf("Unexpected error for pack generation: %v", err)
	}

	// Verify pack directory was created
	expectedPackDir := filepath.Join(tempDir, "test-server-name-1.0.0-npm")
	assertFileExists(t, expectedPackDir)

	// Verify key files were created
	assertFileExists(t, filepath.Join(expectedPackDir, "metadata.hcl"))
	assertFileExists(t, filepath.Join(expectedPackDir, "variables.hcl"))
	assertFileExists(t, filepath.Join(expectedPackDir, "README.md"))
	assertFileExists(t, filepath.Join(expectedPackDir, "outputs.tpl"))
	assertFileExists(t, filepath.Join(expectedPackDir, "templates", "mcp-server.nomad.tpl"))
}

func TestRun_Success_ArchiveGeneration(t *testing.T) {
	ctx := context.Background()
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	opts := Options{
		OutputDir:  tempDir,
		OutputType: "archive",
		DryRun:     false,
	}

	err := Run(ctx, server, serverSpec, "npm", opts)
	if err != nil {
		t.Errorf("Unexpected error for archive generation: %v", err)
	}

	// Verify archive was created
	expectedArchive := filepath.Join(tempDir, "test-server-name-1.0.0-npm.zip")
	assertFileExists(t, expectedArchive)
}

func TestRun_PackdirAlreadyExists(t *testing.T) {
	ctx := context.Background()
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	opts := Options{
		OutputDir:  tempDir,
		OutputType: "packdir",
		DryRun:     false,
	}

	// Create the pack directory first
	expectedPackDir := filepath.Join(tempDir, "test-server-name-1.0.0-npm")
	err := os.MkdirAll(expectedPackDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Try to generate without force flag
	err = Run(ctx, server, serverSpec, "npm", opts)
	if err == nil {
		t.Error("Expected error for existing directory, got none")
	}
	if !contains(err.Error(), "already exists") {
		t.Errorf("Expected error message about existing directory, got: %v", err)
	}
}

func TestRun_PackdirForceOverwrite(t *testing.T) {
	ctx := context.Background()
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	opts := Options{
		OutputDir:  tempDir,
		OutputType: "packdir",
		DryRun:     false,
		Force:      true,
	}

	// Create the pack directory first
	expectedPackDir := filepath.Join(tempDir, "test-server-name-1.0.0-npm")
	err := os.MkdirAll(expectedPackDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Try to generate with force flag
	err = Run(ctx, server, serverSpec, "npm", opts)
	if err != nil {
		t.Errorf("Unexpected error with force flag: %v", err)
	}
}

// Tests for utility functions

func TestSanitizePackName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "name with slashes",
			input:    "org.example/package",
			expected: "org-example-package",
		},
		{
			name:     "name with dots",
			input:    "io.github.user",
			expected: "io-github-user",
		},
		{
			name:     "complex name",
			input:    "io.github.datastax/astra-db-mcp",
			expected: "io-github-datastax-astra-db-mcp",
		},
		{
			name:     "name with special characters",
			input:    "test@#$%^&*()",
			expected: "test",
		},
		{
			name:     "name with spaces",
			input:    "name with spaces",
			expected: "namewithspaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePackName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizePackName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Tests for different package types

func TestRun_DifferentPackageTypes(t *testing.T) {
	tests := []struct {
		name        string
		packageType string
	}{
		{name: "npm package", packageType: "npm"},
		{name: "oci package", packageType: "oci"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			server := mockServerJSON()
			serverSpec := mockServerSpec()
			
			tempDir, cleanup := setupTestDir(t)
			defer cleanup()

			opts := Options{
				OutputDir:  tempDir,
				OutputType: "packdir",
				DryRun:     false,
			}

			err := Run(ctx, server, serverSpec, tt.packageType, opts)
			if err != nil {
				t.Errorf("Unexpected error for %s package: %v", tt.packageType, err)
			}

			// Verify pack was created
			expectedPackDir := filepath.Join(tempDir, "test-server-name-1.0.0-"+tt.packageType)
			assertFileExists(t, expectedPackDir)
		})
	}
}

func TestRun_MinimalServer(t *testing.T) {
	ctx := context.Background()
	server := mockMinimalServerJSON()
	serverSpec := &serversearchutils.ServerSearchSpec{
		Namespace: "minimal",
		Name:      "server",
		Version:    "0.1.0",
	}
	
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	opts := Options{
		OutputDir:  tempDir,
		OutputType: "packdir",
		DryRun:     false,
	}

	err := Run(ctx, server, serverSpec, "npm", opts)
	if err != nil {
		t.Errorf("Unexpected error for minimal server: %v", err)
	}

	// Verify files were created even with minimal data
	expectedPackDir := filepath.Join(tempDir, "minimal-server-0.1.0-npm")
	assertFileExists(t, expectedPackDir)
	assertFileExists(t, filepath.Join(expectedPackDir, "metadata.hcl"))
	assertFileExists(t, filepath.Join(expectedPackDir, "variables.hcl"))
}