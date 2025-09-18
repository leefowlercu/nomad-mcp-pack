package validate

import (
	"fmt"
	"testing"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
)

func TestPackageType(t *testing.T) {
	tests := []struct {
		name          string
		packageType   string
		expectErr     bool
		expectedError string
	}{
		{
			name:        "valid npm",
			packageType: "npm",
			expectErr:   false,
		},
		{
			name:        "valid pypi",
			packageType: "pypi",
			expectErr:   false,
		},
		{
			name:        "valid oci",
			packageType: "oci",
			expectErr:   false,
		},
		{
			name:        "valid nuget",
			packageType: "nuget",
			expectErr:   false,
		},
		{
			name:        "valid uppercase",
			packageType: "NPM",
			expectErr:   false,
		},
		{
			name:        "valid mixed case",
			packageType: "PyPi",
			expectErr:   false,
		},
		{
			name:          "empty package type",
			packageType:   "",
			expectErr:     true,
			expectedError: "invalid package type format; package type must not be empty",
		},
		{
			name:          "invalid package type",
			packageType:   "invalid",
			expectErr:     true,
			expectedError: "invalid package type \"invalid\"; must be one of " + fmt.Sprintf("%v", config.ValidPackageTypes),
		},
		{
			name:          "whitespace only",
			packageType:   "   ",
			expectErr:     true,
			expectedError: "invalid package type \"   \"; must be one of " + fmt.Sprintf("%v", config.ValidPackageTypes),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PackageType(tt.packageType)

			if tt.expectErr {
				if err == nil {
					t.Errorf("PackageType() expected error but got none")
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("PackageType() error = %q, expected %q", err.Error(), tt.expectedError)
				}

			} else {
				if err != nil {
					t.Errorf("PackageType() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestPackageTypes(t *testing.T) {
	tests := []struct {
		name              string
		types             []string
		requireAtLeastOne bool
		expectErr         bool
		errorSubstr       string
	}{
		{
			name:              "empty slice, no requirement",
			types:             []string{},
			requireAtLeastOne: false,
			expectErr:         false,
		},
		{
			name:              "empty slice, require at least one",
			types:             []string{},
			requireAtLeastOne: true,
			expectErr:         true,
			errorSubstr:       "at least one package type must be specified",
		},
		{
			name:              "nil slice, no requirement",
			types:             nil,
			requireAtLeastOne: false,
			expectErr:         false,
		},
		{
			name:              "nil slice, require at least one",
			types:             nil,
			requireAtLeastOne: true,
			expectErr:         true,
			errorSubstr:       "at least one package type must be specified",
		},
		{
			name:              "valid single type",
			types:             []string{"npm"},
			requireAtLeastOne: true,
			expectErr:         false,
		},
		{
			name:              "valid multiple types",
			types:             []string{"npm", "pypi", "oci"},
			requireAtLeastOne: true,
			expectErr:         false,
		},
		{
			name:              "valid types with case variations",
			types:             []string{"NPM", "PyPi", "OCI"},
			requireAtLeastOne: true,
			expectErr:         false,
		},
		{
			name:              "types with whitespace",
			types:             []string{" npm ", "  pypi  "},
			requireAtLeastOne: true,
			expectErr:         false,
		},
		{
			name:              "types with empty strings",
			types:             []string{"npm", "", "pypi"},
			requireAtLeastOne: true,
			expectErr:         false,
		},
		{
			name:              "invalid package type",
			types:             []string{"invalid"},
			requireAtLeastOne: true,
			expectErr:         true,
			errorSubstr:       "invalid package type \"invalid\"",
		},
		{
			name:              "mixed valid and invalid types",
			types:             []string{"npm", "invalid", "pypi"},
			requireAtLeastOne: true,
			expectErr:         true,
			errorSubstr:       "invalid package type \"invalid\"",
		},
		{
			name:              "all empty strings, require at least one",
			types:             []string{"", "   ", ""},
			requireAtLeastOne: true,
			expectErr:         true,
			errorSubstr:       "at least one package type must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PackageTypes(tt.types, tt.requireAtLeastOne)

			if tt.expectErr {
				if err == nil {
					t.Errorf("PackageTypes() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsString(err.Error(), tt.errorSubstr) {
					t.Errorf("PackageTypes() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("PackageTypes() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTransportType(t *testing.T) {
	tests := []struct {
		name          string
		transportType string
		expectErr     bool
		expectedError string
	}{
		{
			name:          "valid stdio",
			transportType: "stdio",
			expectErr:     false,
		},
		{
			name:          "valid http",
			transportType: "http",
			expectErr:     false,
		},
		{
			name:          "valid sse",
			transportType: "sse",
			expectErr:     false,
		},
		{
			name:          "valid uppercase",
			transportType: "HTTP",
			expectErr:     false,
		},
		{
			name:          "valid mixed case",
			transportType: "StDiO",
			expectErr:     false,
		},
		{
			name:          "empty transport type",
			transportType: "",
			expectErr:     true,
			expectedError: "invalid transport type format; transport type must not be empty",
		},
		{
			name:          "invalid transport type",
			transportType: "invalid",
			expectErr:     true,
			expectedError: "invalid transport type \"invalid\"; must be one of " + fmt.Sprintf("%v", config.ValidTransportTypes),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TransportType(tt.transportType)

			if tt.expectErr {
				if err == nil {
					t.Errorf("TransportType() expected error but got none")
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("TransportType() error = %q, expected %q", err.Error(), tt.expectedError)
				}

			} else {
				if err != nil {
					t.Errorf("TransportType() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTransportTypes(t *testing.T) {
	tests := []struct {
		name              string
		types             []string
		requireAtLeastOne bool
		expectErr         bool
		errorSubstr       string
	}{
		{
			name:              "empty slice, no requirement",
			types:             []string{},
			requireAtLeastOne: false,
			expectErr:         false,
		},
		{
			name:              "empty slice, require at least one",
			types:             []string{},
			requireAtLeastOne: true,
			expectErr:         true,
			errorSubstr:       "at least one transport type must be specified",
		},
		{
			name:              "valid single type",
			types:             []string{"stdio"},
			requireAtLeastOne: true,
			expectErr:         false,
		},
		{
			name:              "valid multiple types",
			types:             []string{"stdio", "http", "sse"},
			requireAtLeastOne: true,
			expectErr:         false,
		},
		{
			name:              "invalid transport type",
			types:             []string{"invalid"},
			requireAtLeastOne: true,
			expectErr:         true,
			errorSubstr:       "invalid transport type \"invalid\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TransportTypes(tt.types, tt.requireAtLeastOne)

			if tt.expectErr {
				if err == nil {
					t.Errorf("TransportTypes() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsString(err.Error(), tt.errorSubstr) {
					t.Errorf("TransportTypes() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("TransportTypes() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestOutputDir(t *testing.T) {
	tests := []struct {
		name          string
		outputDir     string
		expectErr     bool
		expectedError string
	}{
		{
			name:      "valid directory",
			outputDir: "/tmp/output",
			expectErr: false,
		},
		{
			name:      "relative path",
			outputDir: "./output",
			expectErr: false,
		},
		{
			name:      "current directory",
			outputDir: ".",
			expectErr: false,
		},
		{
			name:          "empty directory",
			outputDir:     "",
			expectErr:     true,
			expectedError: "invalid output directory format; output directory must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := OutputDir(tt.outputDir)

			if tt.expectErr {
				if err == nil {
					t.Errorf("OutputDir() expected error but got none")
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("OutputDir() error = %q, expected %q", err.Error(), tt.expectedError)
				}

			} else {
				if err != nil {
					t.Errorf("OutputDir() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestOutputType(t *testing.T) {
	tests := []struct {
		name          string
		outputType    string
		expectErr     bool
		expectedError string
	}{
		{
			name:       "valid packdir",
			outputType: "packdir",
			expectErr:  false,
		},
		{
			name:       "valid archive",
			outputType: "archive",
			expectErr:  false,
		},
		{
			name:       "valid uppercase",
			outputType: "PACKDIR",
			expectErr:  false,
		},
		{
			name:       "valid mixed case",
			outputType: "ArChIvE",
			expectErr:  false,
		},
		{
			name:          "empty output type",
			outputType:    "",
			expectErr:     true,
			expectedError: "invalid output type format; output type must not be empty",
		},
		{
			name:          "invalid output type",
			outputType:    "invalid",
			expectErr:     true,
			expectedError: "invalid output type \"invalid\"; must be one of " + fmt.Sprintf("%v", config.ValidOutputTypes),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := OutputType(tt.outputType)

			if tt.expectErr {
				if err == nil {
					t.Errorf("OutputType() expected error but got none")
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("OutputType() error = %q, expected %q", err.Error(), tt.expectedError)
				}

			} else {
				if err != nil {
					t.Errorf("OutputType() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestServerNames(t *testing.T) {
	tests := []struct {
		name        string
		names       []string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "empty slice",
			names:       []string{},
			expectError: false,
		},
		{
			name:        "nil slice",
			names:       nil,
			expectError: false,
		},
		{
			name:        "valid single name",
			names:       []string{"io.github.example/server"},
			expectError: false,
		},
		{
			name:        "valid multiple names",
			names:       []string{"io.github.example/server", "com.example/another"},
			expectError: false,
		},
		{
			name:        "names with whitespace",
			names:       []string{" io.github.example/server ", "  com.example/another  "},
			expectError: false,
		},
		{
			name:        "names with empty strings",
			names:       []string{"io.github.example/server", "", "com.example/another"},
			expectError: false,
		},
		{
			name:        "invalid name format - no slash",
			names:       []string{"invalid-name"},
			expectError: true,
			errorSubstr: "expected format like",
		},
		{
			name:        "invalid name format - empty namespace",
			names:       []string{"/server"},
			expectError: true,
			errorSubstr: "namespace stanza must not be empty",
		},
		{
			name:        "invalid name format - empty name",
			names:       []string{"namespace/"},
			expectError: true,
			errorSubstr: "name stanza must not be empty",
		},
		{
			name:        "mixed valid and invalid names",
			names:       []string{"io.github.example/server", "invalid-name"},
			expectError: true,
			errorSubstr: "expected format like",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ServerNames(tt.names)

			if tt.expectError {
				if err == nil {
					t.Errorf("ServerNames() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsString(err.Error(), tt.errorSubstr) {
					t.Errorf("ServerNames() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("ServerNames() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestPollInterval(t *testing.T) {
	tests := []struct {
		name        string
		interval    int
		expectError bool
		errorSubstr string
	}{
		{
			name:        "valid minimum interval",
			interval:    30,
			expectError: false,
		},
		{
			name:        "valid large interval",
			interval:    3600,
			expectError: false,
		},
		{
			name:        "interval too low",
			interval:    29,
			expectError: true,
			errorSubstr: "poll interval must be at least 30 seconds",
		},
		{
			name:        "zero interval",
			interval:    0,
			expectError: true,
			errorSubstr: "poll interval must be at least 30 seconds",
		},
		{
			name:        "negative interval",
			interval:    -1,
			expectError: true,
			errorSubstr: "poll interval must be at least 30 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PollInterval(tt.interval)

			if tt.expectError {
				if err == nil {
					t.Errorf("PollInterval() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsString(err.Error(), tt.errorSubstr) {
					t.Errorf("PollInterval() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("PollInterval() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestStateFile(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "valid path",
			path:        "/tmp/state.json",
			expectError: false,
		},
		{
			name:        "relative path",
			path:        "./state.json",
			expectError: false,
		},
		{
			name:        "simple filename",
			path:        "state.json",
			expectError: false,
		},
		{
			name:        "empty path",
			path:        "",
			expectError: true,
			errorSubstr: "state file path cannot be empty",
		},
		{
			name:        "whitespace only path",
			path:        "   ",
			expectError: true,
			errorSubstr: "state file path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StateFile(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("StateFile() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsString(err.Error(), tt.errorSubstr) {
					t.Errorf("StateFile() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("StateFile() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestMaxConcurrent(t *testing.T) {
	tests := []struct {
		name        string
		max         int
		expectError bool
		errorSubstr string
	}{
		{
			name:        "valid minimum",
			max:         1,
			expectError: false,
		},
		{
			name:        "valid large number",
			max:         100,
			expectError: false,
		},
		{
			name:        "zero",
			max:         0,
			expectError: true,
			errorSubstr: "max concurrent must be at least 1",
		},
		{
			name:        "negative",
			max:         -1,
			expectError: true,
			errorSubstr: "max concurrent must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MaxConcurrent(tt.max)

			if tt.expectError {
				if err == nil {
					t.Errorf("MaxConcurrent() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !containsString(err.Error(), tt.errorSubstr) {
					t.Errorf("MaxConcurrent() error = %q, expected to contain %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("MaxConcurrent() unexpected error = %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(substr) == 0 ||
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
