package genutils

import (
	"strings"
	"testing"
)

func TestParseServerSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		want        *ServerSpec
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid standard specification",
			spec: "io.github.datastax/astra-db-mcp@0.0.1-seed",
			want: &ServerSpec{
				ServerName: "io.github.datastax/astra-db-mcp",
				Version:    "0.0.1-seed",
			},
			expectError: false,
		},
		{
			name: "valid latest version",
			spec: "example.com/server@latest",
			want: &ServerSpec{
				ServerName: "example.com/server",
				Version:    "latest",
			},
			expectError: false,
		},
		{
			name: "valid complex namespace",
			spec: "org.example.sub/my-server@1.2.3",
			want: &ServerSpec{
				ServerName: "org.example.sub/my-server",
				Version:    "1.2.3",
			},
			expectError: false,
		},
		{
			name: "valid prerelease version",
			spec: "server.io/name@1.0.0-beta.1",
			want: &ServerSpec{
				ServerName: "server.io/name",
				Version:    "1.0.0-beta.1",
			},
			expectError: false,
		},
		{
			name: "specification with whitespace",
			spec: " example.com/server @ 1.0.0 ",
			want: &ServerSpec{
				ServerName: "example.com/server",
				Version:    "1.0.0",
			},
			expectError: false,
		},
		{
			name:        "empty specification",
			spec:        "",
			want:        nil,
			expectError: true,
			errorSubstr: "server specification cannot be empty",
		},
		{
			name:        "missing @ separator",
			spec:        "example.com/server",
			want:        nil,
			expectError: true,
			errorSubstr: "invalid server specification format",
		},
		{
			name:        "multiple @ separators",
			spec:        "example.com/server@1.0.0@extra",
			want:        nil,
			expectError: true,
			errorSubstr: "invalid server specification format",
		},
		{
			name:        "empty server name",
			spec:        "@1.0.0",
			want:        nil,
			expectError: true,
			errorSubstr: "server name cannot be empty",
		},
		{
			name:        "empty version",
			spec:        "example.com/server@",
			want:        nil,
			expectError: true,
			errorSubstr: "version cannot be empty",
		},
		{
			name:        "server name without namespace separator",
			spec:        "server@1.0.0",
			want:        nil,
			expectError: true,
			errorSubstr: "invalid server name format",
		},
		{
			name:        "server name with multiple slashes",
			spec:        "example.com/path/server@1.0.0",
			want:        nil,
			expectError: true,
			errorSubstr: "expected exactly one '/' separator",
		},
		{
			name:        "empty namespace part",
			spec:        "/server@1.0.0",
			want:        nil,
			expectError: true,
			errorSubstr: "namespace and name parts cannot be empty",
		},
		{
			name:        "empty name part",
			spec:        "example.com/@1.0.0",
			want:        nil,
			expectError: true,
			errorSubstr: "namespace and name parts cannot be empty",
		},
		{
			name:        "whitespace only server name",
			spec:        "   @1.0.0",
			want:        nil,
			expectError: true,
			errorSubstr: "server name cannot be empty",
		},
		{
			name:        "whitespace only version",
			spec:        "example.com/server@   ",
			want:        nil,
			expectError: true,
			errorSubstr: "version cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseServerSpec(tt.spec)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseServerSpec() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("ParseServerSpec() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseServerSpec() unexpected error = %v", err)
				return
			}

			if got == nil {
				t.Errorf("ParseServerSpec() returned nil without error")
				return
			}

			if got.ServerName != tt.want.ServerName {
				t.Errorf("ParseServerSpec().ServerName = %q, want %q", got.ServerName, tt.want.ServerName)
			}

			if got.Version != tt.want.Version {
				t.Errorf("ParseServerSpec().Version = %q, want %q", got.Version, tt.want.Version)
			}
		})
	}
}

func TestServerSpec_IsLatest(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "latest lowercase",
			version: "latest",
			want:    true,
		},
		{
			name:    "latest uppercase",
			version: "LATEST",
			want:    true,
		},
		{
			name:    "latest mixed case",
			version: "Latest",
			want:    true,
		},
		{
			name:    "semver version",
			version: "1.0.0",
			want:    false,
		},
		{
			name:    "prerelease version",
			version: "1.0.0-beta.1",
			want:    false,
		},
		{
			name:    "empty version",
			version: "",
			want:    false,
		},
		{
			name:    "latest with extra text",
			version: "latest-version",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ServerSpec{
				ServerName: "example.com/server",
				Version:    tt.version,
			}

			got := spec.IsLatest()
			if got != tt.want {
				t.Errorf("ServerSpec.IsLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerSpec_String(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		version    string
		want       string
	}{
		{
			name:       "standard format",
			serverName: "io.github.datastax/astra-db-mcp",
			version:    "0.0.1-seed",
			want:       "io.github.datastax/astra-db-mcp@0.0.1-seed",
		},
		{
			name:       "latest version",
			serverName: "example.com/server",
			version:    "latest",
			want:       "example.com/server@latest",
		},
		{
			name:       "complex namespace",
			serverName: "org.example.sub/my-server",
			version:    "1.2.3",
			want:       "org.example.sub/my-server@1.2.3",
		},
		{
			name:       "empty values",
			serverName: "",
			version:    "",
			want:       "@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ServerSpec{
				ServerName: tt.serverName,
				Version:    tt.version,
			}

			got := spec.String()
			if got != tt.want {
				t.Errorf("ServerSpec.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseServerSpec_Integration(t *testing.T) {
	validSpecs := []string{
		"io.github.datastax/astra-db-mcp@0.0.1-seed",
		"example.com/server@latest",
		"org.example/test@1.0.0",
	}

	for _, spec := range validSpecs {
		t.Run(spec, func(t *testing.T) {
			parsed, err := ParseServerSpec(spec)
			if err != nil {
				t.Fatalf("ParseServerSpec(%q) unexpected error: %v", spec, err)
			}

			roundTrip := parsed.String()
			if roundTrip != spec {
				t.Errorf("Round trip failed: %q -> %q -> %q", spec, parsed, roundTrip)
			}

			reparsed, err := ParseServerSpec(roundTrip)
			if err != nil {
				t.Fatalf("Re-parsing failed for %q: %v", roundTrip, err)
			}

			if reparsed.ServerName != parsed.ServerName || reparsed.Version != parsed.Version {
				t.Errorf("Re-parsed values differ: original=%+v, reparsed=%+v", parsed, reparsed)
			}
		})
	}
}