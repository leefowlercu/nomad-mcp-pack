package watchutils

import (
	"strings"
	"testing"
)

func TestParseFilterNames(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        []string
		expectError bool
		errorSubstr string
	}{
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
		{
			name:  "whitespace only input",
			input: "   ",
			want:  []string{},
		},
		{
			name:  "single valid server name",
			input: "io.github.example/server",
			want:  []string{"io.github.example/server"},
		},
		{
			name:  "multiple valid server names",
			input: "io.github.example/server,com.example/another,org.test/third",
			want:  []string{"io.github.example/server", "com.example/another", "org.test/third"},
		},
		{
			name:  "server names with whitespace",
			input: " io.github.example/server , com.example/another ",
			want:  []string{"io.github.example/server", "com.example/another"},
		},
		{
			name:  "server names with duplicates",
			input: "io.github.example/server,com.example/another,io.github.example/server",
			want:  []string{"io.github.example/server", "com.example/another"},
		},
		{
			name:  "server names with empty parts",
			input: "io.github.example/server,,com.example/another",
			want:  []string{"io.github.example/server", "com.example/another"},
		},
		{
			name:        "invalid server name format - no slash",
			input:       "invalid-server",
			want:        nil,
			expectError: true,
			errorSubstr: "invalid server name format",
		},
		{
			name:        "invalid server name format - multiple slashes",
			input:       "io.github.example/path/server",
			want:        nil,
			expectError: true,
			errorSubstr: "expected exactly one '/' separator",
		},
		{
			name:        "invalid server name format - empty namespace",
			input:       "/server",
			want:        nil,
			expectError: true,
			errorSubstr: "namespace and name parts cannot be empty",
		},
		{
			name:        "invalid server name format - empty name",
			input:       "namespace/",
			want:        nil,
			expectError: true,
			errorSubstr: "namespace and name parts cannot be empty",
		},
		{
			name:        "mixed valid and invalid names",
			input:       "io.github.example/server,invalid-name",
			want:        nil,
			expectError: true,
			errorSubstr: "invalid server name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilterNames(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseFilterNames() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("ParseFilterNames() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFilterNames() unexpected error = %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseFilterNames() got %d names, want %d", len(got), len(tt.want))
				return
			}

			for i, name := range got {
				if name != tt.want[i] {
					t.Errorf("ParseFilterNames() got[%d] = %q, want %q", i, name, tt.want[i])
				}
			}
		})
	}
}

func TestParsePackageTypes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        []string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "empty input",
			input:       "",
			want:        nil,
			expectError: true,
			errorSubstr: "at least one package type must be specified",
		},
		{
			name:        "whitespace only input",
			input:       "   ",
			want:        nil,
			expectError: true,
			errorSubstr: "at least one package type must be specified",
		},
		{
			name:  "single valid package type",
			input: "npm",
			want:  []string{"npm"},
		},
		{
			name:  "multiple valid package types",
			input: "npm,pypi,oci,nuget",
			want:  []string{"npm", "pypi", "oci", "nuget"},
		},
		{
			name:  "package types with whitespace",
			input: " npm , pypi , oci ",
			want:  []string{"npm", "pypi", "oci"},
		},
		{
			name:  "package types with duplicates",
			input: "npm,pypi,npm,oci",
			want:  []string{"npm", "pypi", "oci"},
		},
		{
			name:  "package types with empty parts",
			input: "npm,,pypi",
			want:  []string{"npm", "pypi"},
		},
		{
			name:  "package types case insensitive",
			input: "NPM,PyPi,OCI,NUGET",
			want:  []string{"npm", "pypi", "oci", "nuget"},
		},
		{
			name:        "invalid package type",
			input:       "invalid",
			want:        nil,
			expectError: true,
			errorSubstr: "invalid package type",
		},
		{
			name:        "mixed valid and invalid package types",
			input:       "npm,invalid,pypi",
			want:        nil,
			expectError: true,
			errorSubstr: "invalid package type",
		},
		{
			name:        "only empty parts after splitting",
			input:       ",,",
			want:        nil,
			expectError: true,
			errorSubstr: "at least one package type must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePackageTypes(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParsePackageTypes() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("ParsePackageTypes() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParsePackageTypes() unexpected error = %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParsePackageTypes() got %d types, want %d", len(got), len(tt.want))
				return
			}

			for i, packageType := range got {
				if packageType != tt.want[i] {
					t.Errorf("ParsePackageTypes() got[%d] = %q, want %q", i, packageType, tt.want[i])
				}
			}
		})
	}
}

func TestValidateWatchConfig(t *testing.T) {
	tests := []struct {
		name          string
		pollInterval  int
		stateFile     string
		maxConcurrent int
		expectError   bool
		errorSubstr   string
	}{
		{
			name:          "valid config",
			pollInterval:  300,
			stateFile:     "watch.json",
			maxConcurrent: 5,
			expectError:   false,
		},
		{
			name:          "minimum poll interval",
			pollInterval:  30,
			stateFile:     "watch.json",
			maxConcurrent: 1,
			expectError:   false,
		},
		{
			name:          "poll interval too low",
			pollInterval:  29,
			stateFile:     "watch.json",
			maxConcurrent: 5,
			expectError:   true,
			errorSubstr:   "poll interval must be at least 30 seconds",
		},
		{
			name:          "empty state file",
			pollInterval:  300,
			stateFile:     "",
			maxConcurrent: 5,
			expectError:   true,
			errorSubstr:   "state file path cannot be empty",
		},
		{
			name:          "whitespace only state file",
			pollInterval:  300,
			stateFile:     "   ",
			maxConcurrent: 5,
			expectError:   true,
			errorSubstr:   "state file path cannot be empty",
		},
		{
			name:          "max concurrent too low",
			pollInterval:  300,
			stateFile:     "watch.json",
			maxConcurrent: 0,
			expectError:   true,
			errorSubstr:   "max concurrent must be at least 1",
		},
		{
			name:          "negative max concurrent",
			pollInterval:  300,
			stateFile:     "watch.json",
			maxConcurrent: -1,
			expectError:   true,
			errorSubstr:   "max concurrent must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWatchConfig(tt.pollInterval, tt.stateFile, tt.maxConcurrent)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateWatchConfig() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("ValidateWatchConfig() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateWatchConfig() unexpected error = %v", err)
			}
		})
	}
}

func TestServerNameFilter_Matches(t *testing.T) {
	tests := []struct {
		name       string
		filter     ServerNameFilter
		serverName string
		want       bool
	}{
		{
			name:       "empty filter matches all",
			filter:     ServerNameFilter{Names: []string{}},
			serverName: "io.github.example/server",
			want:       true,
		},
		{
			name:       "nil filter matches all",
			filter:     ServerNameFilter{Names: nil},
			serverName: "io.github.example/server",
			want:       true,
		},
		{
			name:       "filter matches exact name",
			filter:     ServerNameFilter{Names: []string{"io.github.example/server"}},
			serverName: "io.github.example/server",
			want:       true,
		},
		{
			name:       "filter does not match different name",
			filter:     ServerNameFilter{Names: []string{"io.github.example/server"}},
			serverName: "com.example/other",
			want:       false,
		},
		{
			name:       "filter matches one of multiple names",
			filter:     ServerNameFilter{Names: []string{"io.github.example/server", "com.example/other"}},
			serverName: "com.example/other",
			want:       true,
		},
		{
			name:       "filter does not match any of multiple names",
			filter:     ServerNameFilter{Names: []string{"io.github.example/server", "com.example/other"}},
			serverName: "org.test/unmatched",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.serverName)
			if got != tt.want {
				t.Errorf("ServerNameFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPackageTypeFilter_Matches(t *testing.T) {
	tests := []struct {
		name        string
		filter      PackageTypeFilter
		packageType string
		want        bool
	}{
		{
			name:        "filter matches exact type",
			filter:      PackageTypeFilter{Types: []string{"npm"}},
			packageType: "npm",
			want:        true,
		},
		{
			name:        "filter matches case insensitive",
			filter:      PackageTypeFilter{Types: []string{"npm"}},
			packageType: "NPM",
			want:        true,
		},
		{
			name:        "filter does not match different type",
			filter:      PackageTypeFilter{Types: []string{"npm"}},
			packageType: "pypi",
			want:        false,
		},
		{
			name:        "filter matches one of multiple types",
			filter:      PackageTypeFilter{Types: []string{"npm", "pypi", "oci"}},
			packageType: "pypi",
			want:        true,
		},
		{
			name:        "filter does not match any of multiple types",
			filter:      PackageTypeFilter{Types: []string{"npm", "pypi"}},
			packageType: "nuget",
			want:        false,
		},
		{
			name:        "empty filter does not match",
			filter:      PackageTypeFilter{Types: []string{}},
			packageType: "npm",
			want:        false,
		},
		{
			name:        "nil filter does not match",
			filter:      PackageTypeFilter{Types: nil},
			packageType: "npm",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.packageType)
			if got != tt.want {
				t.Errorf("PackageTypeFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidPackageType(t *testing.T) {
	tests := []struct {
		name        string
		packageType string
		want        bool
	}{
		{
			name:        "valid npm",
			packageType: "npm",
			want:        true,
		},
		{
			name:        "valid pypi",
			packageType: "pypi",
			want:        true,
		},
		{
			name:        "valid oci",
			packageType: "oci",
			want:        true,
		},
		{
			name:        "valid nuget",
			packageType: "nuget",
			want:        true,
		},
		{
			name:        "valid uppercase",
			packageType: "NPM",
			want:        true,
		},
		{
			name:        "valid mixed case",
			packageType: "PyPi",
			want:        true,
		},
		{
			name:        "invalid type",
			packageType: "invalid",
			want:        false,
		},
		{
			name:        "empty string",
			packageType: "",
			want:        false,
		},
		{
			name:        "similar but wrong",
			packageType: "docker",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPackageType(tt.packageType)
			if got != tt.want {
				t.Errorf("isValidPackageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidPackageTypes(t *testing.T) {
	// Test that our constant contains the expected values
	expected := []string{"npm", "pypi", "oci", "nuget"}
	
	if len(ValidPackageTypes) != len(expected) {
		t.Errorf("ValidPackageTypes length = %d, want %d", len(ValidPackageTypes), len(expected))
	}

	for i, packageType := range expected {
		if ValidPackageTypes[i] != packageType {
			t.Errorf("ValidPackageTypes[%d] = %q, want %q", i, ValidPackageTypes[i], packageType)
		}
	}
}