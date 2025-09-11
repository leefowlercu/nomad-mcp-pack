package serversearchutils

import (
	"strings"
	"testing"
)

func TestParseServerSearchSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		want        *ServerSearchSpec
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid standard specification",
			spec: "io.github.datastax/astra-db-mcp@0.0.1-seed",
			want: &ServerSearchSpec{
				Namespace: "io.github.datastax",
				Name:      "astra-db-mcp",
				Version:   "0.0.1-seed",
			},
			expectError: false,
		},
		{
			name: "valid latest version",
			spec: "example.com/server@latest",
			want: &ServerSearchSpec{
				Namespace: "example.com",
				Name:      "server",
				Version:   "latest",
			},
			expectError: false,
		},
		{
			name: "valid complex namespace",
			spec: "org.example.sub/my-server@1.2.3",
			want: &ServerSearchSpec{
				Namespace: "org.example.sub",
				Name:      "my-server",
				Version:   "1.2.3",
			},
			expectError: false,
		},
		{
			name: "valid prerelease version",
			spec: "server.io/name@1.0.0-beta.1",
			want: &ServerSearchSpec{
				Namespace: "server.io",
				Name:      "name",
				Version:   "1.0.0-beta.1",
			},
			expectError: false,
		},
		{
			name: "specification with whitespace",
			spec: " example.com/server @ 1.0.0 ",
			want: &ServerSearchSpec{
				Namespace: "example.com",
				Name:      "server",
				Version:   "1.0.0",
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
			got, err := ParseServerSearchSpec(tt.spec)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseServerSearchSpec() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("ParseServerSearchSpec() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseServerSearchSpec() unexpected error = %v", err)
				return
			}

			if got == nil {
				t.Errorf("ParseServerSearchSpec() returned nil without error")
				return
			}

			if got.Namespace != tt.want.Namespace {
				t.Errorf("ParseServerSearchSpec().Namespace = %q, want %q", got.Namespace, tt.want.Namespace)
			}

			if got.Name != tt.want.Name {
				t.Errorf("ParseServerSearchSpec().Name = %q, want %q", got.Name, tt.want.Name)
			}

			if got.Version != tt.want.Version {
				t.Errorf("ParseServerSearchSpec().Version = %q, want %q", got.Version, tt.want.Version)
			}

			// Test that FullName() returns the expected server name
			expectedFullName := tt.want.Namespace + "/" + tt.want.Name
			if got.FullName() != expectedFullName {
				t.Errorf("ParseServerSearchSpec().FullName() = %q, want %q", got.FullName(), expectedFullName)
			}
		})
	}
}

func TestServerSearchSpec_IsLatest(t *testing.T) {
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
			spec := &ServerSearchSpec{
				Namespace: "example.com",
				Name:      "server",
				Version:   tt.version,
			}

			got := spec.IsLatest()
			if got != tt.want {
				t.Errorf("ServerSearchSpec.IsLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerSearchSpec_String(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		serverName string
		version   string
		want      string
	}{
		{
			name:       "standard format",
			namespace:  "io.github.datastax",
			serverName: "astra-db-mcp",
			version:    "0.0.1-seed",
			want:       "io.github.datastax/astra-db-mcp@0.0.1-seed",
		},
		{
			name:       "latest version",
			namespace:  "example.com",
			serverName: "server",
			version:    "latest",
			want:       "example.com/server@latest",
		},
		{
			name:       "complex namespace",
			namespace:  "org.example.sub",
			serverName: "my-server",
			version:    "1.2.3",
			want:       "org.example.sub/my-server@1.2.3",
		},
		{
			name:       "empty values",
			namespace:  "",
			serverName: "",
			version:    "",
			want:       "/@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ServerSearchSpec{
				Namespace: tt.namespace,
				Name:      tt.serverName,
				Version:   tt.version,
			}

			got := spec.String()
			if got != tt.want {
				t.Errorf("ServerSearchSpec.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestServerSearchSpec_FullName(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		serverName string
		want      string
	}{
		{
			name:       "standard format",
			namespace:  "io.github.datastax",
			serverName: "astra-db-mcp",
			want:       "io.github.datastax/astra-db-mcp",
		},
		{
			name:       "simple format",
			namespace:  "example.com",
			serverName: "server",
			want:       "example.com/server",
		},
		{
			name:       "complex namespace",
			namespace:  "org.example.sub",
			serverName: "my-server",
			want:       "org.example.sub/my-server",
		},
		{
			name:       "empty values",
			namespace:  "",
			serverName: "",
			want:       "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ServerSearchSpec{
				Namespace: tt.namespace,
				Name:      tt.serverName,
				Version:   "1.0.0", // Version shouldn't affect FullName
			}

			got := spec.FullName()
			if got != tt.want {
				t.Errorf("ServerSearchSpec.FullName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateServerName(t *testing.T) {
	tests := []struct {
		name            string
		fullName        string
		wantNamespace   string
		wantServerName  string
		expectError     bool
		errorSubstr     string
	}{
		{
			name:           "valid standard format",
			fullName:       "io.github.datastax/astra-db-mcp",
			wantNamespace:  "io.github.datastax",
			wantServerName: "astra-db-mcp",
			expectError:    false,
		},
		{
			name:           "valid simple format",
			fullName:       "example.com/server",
			wantNamespace:  "example.com",
			wantServerName: "server",
			expectError:    false,
		},
		{
			name:           "valid complex namespace",
			fullName:       "org.example.sub/my-server",
			wantNamespace:  "org.example.sub",
			wantServerName: "my-server",
			expectError:    false,
		},
		{
			name:           "with whitespace",
			fullName:       " example.com/server ",
			wantNamespace:  "example.com",
			wantServerName: "server",
			expectError:    false,
		},
		{
			name:        "empty name",
			fullName:    "",
			expectError: true,
			errorSubstr: "server name cannot be empty",
		},
		{
			name:        "whitespace only",
			fullName:    "   ",
			expectError: true,
			errorSubstr: "server name cannot be empty",
		},
		{
			name:        "missing separator",
			fullName:    "server",
			expectError: true,
			errorSubstr: "invalid server name format",
		},
		{
			name:        "multiple separators",
			fullName:    "example.com/path/server",
			expectError: true,
			errorSubstr: "expected exactly one '/' separator",
		},
		{
			name:        "empty namespace",
			fullName:    "/server",
			expectError: true,
			errorSubstr: "namespace and name parts cannot be empty",
		},
		{
			name:        "empty server name part",
			fullName:    "example.com/",
			expectError: true,
			errorSubstr: "namespace and name parts cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNamespace, gotName, err := ValidateServerName(tt.fullName)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateServerName() expected error but got none")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("ValidateServerName() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateServerName() unexpected error = %v", err)
				return
			}

			if gotNamespace != tt.wantNamespace {
				t.Errorf("ValidateServerName() namespace = %q, want %q", gotNamespace, tt.wantNamespace)
			}

			if gotName != tt.wantServerName {
				t.Errorf("ValidateServerName() name = %q, want %q", gotName, tt.wantServerName)
			}
		})
	}
}

func TestParseServerName(t *testing.T) {
	// ParseServerName is just a wrapper around ValidateServerName, so we test a few key cases
	tests := []struct {
		name       string
		fullName   string
		wantNS     string
		wantName   string
		wantError  bool
	}{
		{
			name:      "valid name",
			fullName:  "io.github.datastax/astra-db-mcp",
			wantNS:    "io.github.datastax",
			wantName:  "astra-db-mcp",
			wantError: false,
		},
		{
			name:      "invalid name",
			fullName:  "invalid-name",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNS, gotName, err := ParseServerName(tt.fullName)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseServerName() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseServerName() unexpected error = %v", err)
				return
			}

			if gotNS != tt.wantNS {
				t.Errorf("ParseServerName() namespace = %q, want %q", gotNS, tt.wantNS)
			}

			if gotName != tt.wantName {
				t.Errorf("ParseServerName() name = %q, want %q", gotName, tt.wantName)
			}
		})
	}
}

func TestParseServerSearchSpec_Integration(t *testing.T) {
	validSpecs := []string{
		"io.github.datastax/astra-db-mcp@0.0.1-seed",
		"example.com/server@latest",
		"org.example/test@1.0.0",
	}

	for _, spec := range validSpecs {
		t.Run(spec, func(t *testing.T) {
			parsed, err := ParseServerSearchSpec(spec)
			if err != nil {
				t.Fatalf("ParseServerSearchSpec(%q) unexpected error: %v", spec, err)
			}

			roundTrip := parsed.String()
			if roundTrip != spec {
				t.Errorf("Round trip failed: %q -> %q -> %q", spec, parsed, roundTrip)
			}

			reparsed, err := ParseServerSearchSpec(roundTrip)
			if err != nil {
				t.Fatalf("Re-parsing failed for %q: %v", roundTrip, err)
			}

			if reparsed.Namespace != parsed.Namespace || reparsed.Name != parsed.Name || reparsed.Version != parsed.Version {
				t.Errorf("Re-parsed values differ: original=%+v, reparsed=%+v", parsed, reparsed)
			}
		})
	}
}