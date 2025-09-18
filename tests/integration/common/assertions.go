package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AssertCommandSuccess asserts that a CLI command succeeded
func AssertCommandSuccess(t *testing.T, result *CommandResult, msgAndArgs ...interface{}) {
	t.Helper()

	if !result.IsSuccess() {
		msg := fmt.Sprintf("Command failed with exit code %d", result.ExitCode)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}

		t.Errorf("%s\nCommand: %s %s\nStdout: %s\nStderr: %s\nError: %v",
			msg, result.Command, strings.Join(result.Args, " "),
			result.Stdout, result.Stderr, result.Error)
	}
}

// AssertCommandFailure asserts that a CLI command failed
func AssertCommandFailure(t *testing.T, result *CommandResult, msgAndArgs ...interface{}) {
	t.Helper()

	if result.IsSuccess() {
		msg := "Expected command to fail but it succeeded"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}

		t.Errorf("%s\nCommand: %s %s\nStdout: %s\nStderr: %s",
			msg, result.Command, strings.Join(result.Args, " "),
			result.Stdout, result.Stderr)
	}
}

// AssertErrorContains asserts that the command error contains specific text
func AssertErrorContains(t *testing.T, result *CommandResult, expectedText string, msgAndArgs ...interface{}) {
	t.Helper()

	output := result.GetOutput()
	if !strings.Contains(output, expectedText) {
		msg := fmt.Sprintf("Expected error output to contain %q", expectedText)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}

		t.Errorf("%s\nActual output: %s", msg, output)
	}
}

// AssertOutputContains asserts that the command output contains specific text
func AssertOutputContains(t *testing.T, result *CommandResult, expectedText string, msgAndArgs ...interface{}) {
	t.Helper()

	output := result.GetOutput()
	if !strings.Contains(output, expectedText) {
		msg := fmt.Sprintf("Expected output to contain %q", expectedText)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}

		t.Errorf("%s\nActual output: %s", msg, output)
	}
}

// AssertPackGenerated asserts that a pack was generated in the specified directory
func AssertPackGenerated(t *testing.T, outputDir, packName string, msgAndArgs ...interface{}) {
	t.Helper()

	packDir := filepath.Join(outputDir, packName)
	if _, err := os.Stat(packDir); os.IsNotExist(err) {
		msg := fmt.Sprintf("Pack directory %s does not exist", packDir)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Error(msg)
		return
	}

	// Verify it's a directory
	info, err := os.Stat(packDir)
	if err != nil {
		t.Errorf("Failed to stat pack directory %s: %v", packDir, err)
		return
	}

	if !info.IsDir() {
		t.Errorf("Pack path %s is not a directory", packDir)
	}
}

// AssertArchiveGenerated asserts that a pack archive was generated
func AssertArchiveGenerated(t *testing.T, outputDir, packName string, msgAndArgs ...interface{}) {
	t.Helper()

	archivePath := filepath.Join(outputDir, packName+".zip")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		msg := fmt.Sprintf("Pack archive %s does not exist", archivePath)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Error(msg)
		return
	}

	// Verify it's a file
	info, err := os.Stat(archivePath)
	if err != nil {
		t.Errorf("Failed to stat pack archive %s: %v", archivePath, err)
		return
	}

	if info.IsDir() {
		t.Errorf("Pack archive path %s is a directory, expected file", archivePath)
	}
}

// AssertPackStructure asserts that a pack has the expected file structure
func AssertPackStructure(t *testing.T, packDir string, msgAndArgs ...interface{}) {
	t.Helper()

	requiredFiles := []string{
		"metadata.hcl",
		"variables.hcl",
		"outputs.tpl",
		"README.md",
		"templates/mcp-server.nomad.tpl",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(packDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			msg := fmt.Sprintf("Required file %s does not exist in pack", file)
			if len(msgAndArgs) > 0 {
				msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
			}
			t.Error(msg)
		}
	}
}

// AssertFileExists asserts that a file exists
func AssertFileExists(t *testing.T, filePath string, msgAndArgs ...interface{}) {
	t.Helper()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		msg := fmt.Sprintf("File %s does not exist", filePath)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Error(msg)
	}
}

// AssertFileNotExists asserts that a file does not exist
func AssertFileNotExists(t *testing.T, filePath string, msgAndArgs ...interface{}) {
	t.Helper()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		msg := fmt.Sprintf("File %s should not exist", filePath)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Error(msg)
	}
}

// AssertFileContains asserts that a file contains specific content
func AssertFileContains(t *testing.T, filePath, expectedContent string, msgAndArgs ...interface{}) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", filePath, err)
		return
	}

	if !strings.Contains(string(content), expectedContent) {
		msg := fmt.Sprintf("File %s does not contain expected content %q", filePath, expectedContent)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Errorf("%s\nActual content:\n%s", msg, string(content))
	}
}

// AssertFileNotContains asserts that a file does not contain specific content
func AssertFileNotContains(t *testing.T, filePath, unexpectedContent string, msgAndArgs ...interface{}) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", filePath, err)
		return
	}

	if strings.Contains(string(content), unexpectedContent) {
		msg := fmt.Sprintf("File %s contains unexpected content %q", filePath, unexpectedContent)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Errorf("%s\nActual content:\n%s", msg, string(content))
	}
}

// AssertDirExists asserts that a directory exists
func AssertDirExists(t *testing.T, dirPath string, msgAndArgs ...interface{}) {
	t.Helper()

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		msg := fmt.Sprintf("Directory %s does not exist", dirPath)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Error(msg)
		return
	}

	if err != nil {
		t.Errorf("Failed to stat directory %s: %v", dirPath, err)
		return
	}

	if !info.IsDir() {
		msg := fmt.Sprintf("Path %s is not a directory", dirPath)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Error(msg)
	}
}

// AssertDirEmpty asserts that a directory is empty
func AssertDirEmpty(t *testing.T, dirPath string, msgAndArgs ...interface{}) {
	t.Helper()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		t.Errorf("Failed to read directory %s: %v", dirPath, err)
		return
	}

	if len(entries) > 0 {
		msg := fmt.Sprintf("Directory %s is not empty (contains %d entries)", dirPath, len(entries))
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Error(msg)
	}
}

// AssertDirNotEmpty asserts that a directory is not empty
func AssertDirNotEmpty(t *testing.T, dirPath string, msgAndArgs ...interface{}) {
	t.Helper()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		t.Errorf("Failed to read directory %s: %v", dirPath, err)
		return
	}

	if len(entries) == 0 {
		msg := fmt.Sprintf("Directory %s is empty", dirPath)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%v: %s", msgAndArgs[0], msg)
		}
		t.Error(msg)
	}
}

// ComputePackName computes the expected pack name for a server
func ComputePackName(serverName, version, packageType, transportType string) string {
	// Sanitize server name (same logic as in the generator)
	sanitized := strings.ReplaceAll(serverName, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-")

	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	return fmt.Sprintf("%s-%s-%s-%s", result.String(), version, packageType, transportType)
}