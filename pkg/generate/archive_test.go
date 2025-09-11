package generate

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leefowlercu/nomad-mcp-pack/internal/serversearchutils"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

// Helper function to create a test pack directory with files
func createTestPackDir(t *testing.T, baseDir string) string {
	packName := "test-pack"
	packDir := filepath.Join(baseDir, packName)
	
	// Create directory structure
	if err := os.MkdirAll(packDir, 0755); err != nil {
		t.Fatalf("Failed to create pack directory: %v", err)
	}
	
	templatesDir := filepath.Join(packDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}
	
	// Create test files
	testFiles := map[string]string{
		"metadata.hcl":                     "app { url = \"https://test.com\" }\npack { name = \"test\" }",
		"variables.hcl":                    "variable \"test\" { type = string }",
		"outputs.tpl":                      "Output: test successful",
		"README.md":                        "# Test Pack\n\nThis is a test pack.",
		"templates/mcp-server.nomad.tpl":   "job \"test\" { type = \"service\" }",
	}
	
	for relPath, content := range testFiles {
		fullPath := filepath.Join(packDir, relPath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", relPath, err)
		}
	}
	
	return packDir
}

// Helper function to verify ZIP archive contents
func assertArchiveContains(t *testing.T, archivePath string, expectedFiles []string) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("Failed to open archive: %v", err)
	}
	defer reader.Close()

	foundFiles := make(map[string]bool)
	for _, file := range reader.File {
		foundFiles[file.Name] = true
	}

	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("Expected file %s not found in archive", expected)
		}
	}
}

// Helper function to get file count in ZIP archive
func getArchiveFileCount(t *testing.T, archivePath string) int {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("Failed to open archive: %v", err)
	}
	defer reader.Close()

	count := 0
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			count++
		}
	}
	return count
}

func TestCreateArchive_Success(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test pack directory
	packDir := createTestPackDir(t, tempDir)

	// Create generator with test data
	generator := &Generator{
		serverSpec: &serversearchutils.ServerSearchSpec{
			Namespace: "test.example", Name: "server",
			Version:    "1.0.0",
		},
		pkg: &model.Package{
			RegistryType: "npm",
		},
		options: Options{
			OutputDir: tempDir,
			Force:     false,
		},
	}

	// Create archive
	err := generator.createArchive(packDir)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	// Verify archive was created
	expectedArchive := filepath.Join(tempDir, "test-example-server-1.0.0-npm.zip")
	assertFileExists(t, expectedArchive)

	// Verify archive contains expected files
	expectedFiles := []string{
		"metadata.hcl",
		"variables.hcl",
		"outputs.tpl",
		"README.md",
		"templates/mcp-server.nomad.tpl",
	}
	assertArchiveContains(t, expectedArchive, expectedFiles)

	// Verify correct number of files (5 files + 1 directory entry)
	fileCount := getArchiveFileCount(t, expectedArchive)
	if fileCount != 5 {
		t.Errorf("Expected 5 files in archive, got %d", fileCount)
	}
}

func TestCreateArchive_AlreadyExists(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test pack directory
	packDir := createTestPackDir(t, tempDir)

	// Create generator
	generator := &Generator{
		serverSpec: &serversearchutils.ServerSearchSpec{
			Namespace: "test", Name: "server",
			Version:    "1.0.0",
		},
		pkg: &model.Package{
			RegistryType: "npm",
		},
		options: Options{
			OutputDir: tempDir,
			Force:     false,
		},
	}

	// Create an existing archive file
	expectedArchive := filepath.Join(tempDir, "test-server-1.0.0-npm.zip")
	if err := os.WriteFile(expectedArchive, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing archive: %v", err)
	}

	// Try to create archive without force flag
	err := generator.createArchive(packDir)
	if err == nil {
		t.Error("Expected error for existing archive, got none")
	}
	if !contains(err.Error(), "already exists") {
		t.Errorf("Expected error about existing archive, got: %v", err)
	}
}

func TestCreateArchive_ForceOverwrite(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test pack directory
	packDir := createTestPackDir(t, tempDir)

	// Create generator with force flag
	generator := &Generator{
		serverSpec: &serversearchutils.ServerSearchSpec{
			Namespace: "test", Name: "server",
			Version:    "1.0.0",
		},
		pkg: &model.Package{
			RegistryType: "npm",
		},
		options: Options{
			OutputDir: tempDir,
			Force:     true,
		},
	}

	// Create an existing archive file
	expectedArchive := filepath.Join(tempDir, "test-server-1.0.0-npm.zip")
	if err := os.WriteFile(expectedArchive, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing archive: %v", err)
	}

	// Try to create archive with force flag
	err := generator.createArchive(packDir)
	if err != nil {
		t.Errorf("Unexpected error with force flag: %v", err)
	}

	// Verify archive was overwritten (should be larger than "existing")
	info, err := os.Stat(expectedArchive)
	if err != nil {
		t.Fatalf("Failed to stat archive: %v", err)
	}
	if info.Size() <= 8 { // "existing" is 8 bytes
		t.Error("Archive was not overwritten with force flag")
	}
}

func TestCreateArchive_EmptyDirectory(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create empty pack directory
	packDir := filepath.Join(tempDir, "empty-pack")
	if err := os.MkdirAll(packDir, 0755); err != nil {
		t.Fatalf("Failed to create empty pack directory: %v", err)
	}

	// Create generator
	generator := &Generator{
		serverSpec: &serversearchutils.ServerSearchSpec{
			Namespace: "empty", Name: "server",
			Version:    "1.0.0",
		},
		pkg: &model.Package{
			RegistryType: "npm",
		},
		options: Options{
			OutputDir: tempDir,
			Force:     false,
		},
	}

	// Create archive of empty directory
	err := generator.createArchive(packDir)
	if err != nil {
		t.Errorf("Failed to create archive of empty directory: %v", err)
	}

	// Verify archive was created
	expectedArchive := filepath.Join(tempDir, "empty-server-1.0.0-npm.zip")
	assertFileExists(t, expectedArchive)

	// Verify archive has no files (just the root directory)
	fileCount := getArchiveFileCount(t, expectedArchive)
	if fileCount != 0 {
		t.Errorf("Expected 0 files in empty archive, got %d", fileCount)
	}
}

func TestCreateArchive_NestedDirectories(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create pack directory with nested structure
	packName := "nested-pack"
	packDir := filepath.Join(tempDir, packName)
	
	nestedPaths := []string{
		"metadata.hcl",
		"templates/job.nomad.tpl",
		"templates/helpers/_helper.tpl",
		"config/settings.hcl",
	}
	
	for _, path := range nestedPaths {
		fullPath := filepath.Join(packDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Create generator
	generator := &Generator{
		serverSpec: &serversearchutils.ServerSearchSpec{
			Namespace: "nested", Name: "server",
			Version:    "1.0.0",
		},
		pkg: &model.Package{
			RegistryType: "npm",
		},
		options: Options{
			OutputDir: tempDir,
			Force:     false,
		},
	}

	// Create archive
	err := generator.createArchive(packDir)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	// Verify archive contains nested files
	expectedArchive := filepath.Join(tempDir, "nested-server-1.0.0-npm.zip")
	expectedFiles := []string{
		"metadata.hcl",
		"templates/job.nomad.tpl",
		"templates/helpers/_helper.tpl",
		"config/settings.hcl",
	}
	assertArchiveContains(t, expectedArchive, expectedFiles)
}

func TestCreateArchive_InvalidPath(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create generator with invalid output directory
	generator := &Generator{
		serverSpec: &serversearchutils.ServerSearchSpec{
			Namespace: "test", Name: "server",
			Version:    "1.0.0",
		},
		pkg: &model.Package{
			RegistryType: "npm",
		},
		options: Options{
			OutputDir: "/invalid/nonexistent/path",
			Force:     false,
		},
	}

	// Try to create archive with invalid path
	packDir := filepath.Join(tempDir, "test-pack")
	os.MkdirAll(packDir, 0755)
	
	err := generator.createArchive(packDir)
	if err == nil {
		t.Error("Expected error for invalid output directory, got none")
	}
}

func TestCreateArchive_FilePermissions(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create pack directory with executable file
	packName := "perm-pack"
	packDir := filepath.Join(tempDir, packName)
	os.MkdirAll(packDir, 0755)

	// Create files with different permissions
	scriptPath := filepath.Join(packDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	configPath := filepath.Join(packDir, "config.hcl")
	if err := os.WriteFile(configPath, []byte("config = true"), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Create generator
	generator := &Generator{
		serverSpec: &serversearchutils.ServerSearchSpec{
			Namespace: "perm", Name: "server",
			Version:    "1.0.0",
		},
		pkg: &model.Package{
			RegistryType: "npm",
		},
		options: Options{
			OutputDir: tempDir,
			Force:     false,
		},
	}

	// Create archive
	err := generator.createArchive(packDir)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	// Verify archive was created successfully
	expectedArchive := filepath.Join(tempDir, "perm-server-1.0.0-npm.zip")
	assertFileExists(t, expectedArchive)

	// Verify both files are in archive
	expectedFiles := []string{"script.sh", "config.hcl"}
	assertArchiveContains(t, expectedArchive, expectedFiles)
}

func TestCreateArchive_PackNameSanitization(t *testing.T) {
	tests := []struct {
		name           string
		serverName     string
		version        string
		packageType    string
		expectedPrefix string
	}{
		{
			name:           "simple name",
			serverName:     "example.com/simple-server",
			version:        "1.0.0",
			packageType:    "npm",
			expectedPrefix: "example-com-simple-server-1.0.0-npm",
		},
		{
			name:           "name with slashes",
			serverName:     "org.example/server",
			version:        "2.1.0",
			packageType:    "oci",
			expectedPrefix: "org-example-server-2.1.0-oci",
		},
		{
			name:           "name with dots and special chars",
			serverName:     "io.github.user/complex-server",
			version:        "0.1.0-beta",
			packageType:    "pypi",
			expectedPrefix: "io-github-user-complex-server-0.1.0-beta-pypi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := setupTestDir(t)
			defer cleanup()

			packDir := createTestPackDir(t, tempDir)

			// Parse namespace/name format
			parts := strings.Split(tt.serverName, "/")
			serverSpec := &serversearchutils.ServerSearchSpec{
				Namespace: parts[0],
				Name:      parts[1],
				Version:   tt.version,
			}

			generator := &Generator{
				serverSpec: serverSpec,
				pkg: &model.Package{
					RegistryType: tt.packageType,
				},
				options: Options{
					OutputDir: tempDir,
					Force:     false,
				},
			}

			err := generator.createArchive(packDir)
			if err != nil {
				t.Fatalf("Failed to create archive: %v", err)
			}

			// Verify archive name is sanitized correctly
			expectedArchive := filepath.Join(tempDir, tt.expectedPrefix+".zip")
			assertFileExists(t, expectedArchive)
		})
	}
}