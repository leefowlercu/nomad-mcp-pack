package integration_test

import (
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

// Sample test data for integration testing
// These represent expected structures that might be found in a dev registry

// sampleActiveServer returns a sample active server for testing
func sampleActiveServer() v0.ServerJSON {
	return v0.ServerJSON{
		Name:        "sample-test-server",
		Description: "A sample test server for integration testing",
		Version:     "1.0.0",
		Status:      model.StatusActive,
		Repository: model.Repository{
			URL:    "https://github.com/example/sample-test-server",
			Source: "github",
		},
		Packages: []model.Package{
			{
				RegistryType: "npm",
				Identifier:   "sample-test-server",
				Version:      "1.0.0",
			},
		},
	}
}

// sampleDeprecatedServer returns a sample deprecated server for testing
func sampleDeprecatedServer() v0.ServerJSON {
	return v0.ServerJSON{
		Name:        "deprecated-test-server",
		Description: "A deprecated test server",
		Version:     "0.5.0",
		Status:      model.StatusDeprecated,
		Repository: model.Repository{
			URL:    "https://github.com/example/deprecated-test-server",
			Source: "github",
		},
		Packages: []model.Package{
			{
				RegistryType: "npm",
				Identifier:   "deprecated-test-server",
				Version:      "0.5.0",
			},
		},
	}
}

// sampleServerVersions returns multiple versions of the same server
func sampleServerVersions() []v0.ServerJSON {
	return []v0.ServerJSON{
		{
			Name:        "multi-version-server",
			Description: "A server with multiple versions - v1.0.0",
			Version:     "1.0.0",
			Status:      model.StatusActive,
			Repository: model.Repository{
				URL:    "https://github.com/example/multi-version-server",
				Source: "github",
			},
			Packages: []model.Package{
				{
					RegistryType: "npm",
					Identifier:   "multi-version-server",
					Version:      "1.0.0",
				},
			},
		},
		{
			Name:        "multi-version-server",
			Description: "A server with multiple versions - v1.5.0",
			Version:     "1.5.0",
			Status:      model.StatusActive,
			Repository: model.Repository{
				URL:    "https://github.com/example/multi-version-server",
				Source: "github",
			},
			Packages: []model.Package{
				{
					RegistryType: "npm",
					Identifier:   "multi-version-server",
					Version:      "1.5.0",
				},
			},
		},
		{
			Name:        "multi-version-server",
			Description: "A server with multiple versions - v2.0.0",
			Version:     "2.0.0",
			Status:      model.StatusActive,
			Repository: model.Repository{
				URL:    "https://github.com/example/multi-version-server",
				Source: "github",
			},
			Packages: []model.Package{
				{
					RegistryType: "npm",
					Identifier:   "multi-version-server",
					Version:      "2.0.0",
				},
			},
		},
		{
			Name:        "multi-version-server",
			Description: "A server with multiple versions - v3.0.0 (deprecated)",
			Version:     "3.0.0",
			Status:      model.StatusDeprecated,
			Repository: model.Repository{
				URL:    "https://github.com/example/multi-version-server",
				Source: "github",
			},
			Packages: []model.Package{
				{
					RegistryType: "npm",
					Identifier:   "multi-version-server",
					Version:      "3.0.0",
				},
			},
		},
	}
}

// sampleFilesystemServer returns a sample filesystem-related server
func sampleFilesystemServer() v0.ServerJSON {
	return v0.ServerJSON{
		Name:        "filesystem-server",
		Description: "A filesystem management server",
		Version:     "2.1.0",
		Status:      model.StatusActive,
		Repository: model.Repository{
			URL:    "https://github.com/example/filesystem-server",
			Source: "github",
		},
		Packages: []model.Package{
			{
				RegistryType: "pypi",
				Identifier:   "filesystem-server",
				Version:      "2.1.0",
			},
		},
	}
}

// sampleOCIServer returns a sample OCI/Docker-based server
func sampleOCIServer() v0.ServerJSON {
	return v0.ServerJSON{
		Name:        "docker-test-server",
		Description: "A Docker-based test server",
		Version:     "1.2.0",
		Status:      model.StatusActive,
		Repository: model.Repository{
			URL:    "https://github.com/example/docker-test-server",
			Source: "github",
		},
		Packages: []model.Package{
			{
				RegistryType: "oci",
				Identifier:   "docker-test-server",
				Version:      "1.2.0",
			},
		},
	}
}

// sampleNuGetServer returns a sample NuGet-based server
func sampleNuGetServer() v0.ServerJSON {
	return v0.ServerJSON{
		Name:        "dotnet-test-server",
		Description: "A .NET-based test server",
		Version:     "1.0.0",
		Status:      model.StatusActive,
		Repository: model.Repository{
			URL:    "https://github.com/example/dotnet-test-server",
			Source: "github",
		},
		Packages: []model.Package{
			{
				RegistryType: "nuget",
				Identifier:   "dotnet-test-server",
				Version:      "1.0.0",
			},
		},
	}
}

// expectedSearchTerms returns common search terms that should return results
func expectedSearchTerms() []string {
	return []string{
		"file",
		"test",
		"server",
		"sample",
		"filesystem",
		"docker",
		"npm",
	}
}

// expectedPackageTypes returns the package types we expect to support
func expectedPackageTypes() []string {
	return []string{
		"npm",
		"pypi",
		"oci",
		"nuget",
	}
}

// sampleMetadata returns sample metadata for testing
func sampleMetadata() *v0.Metadata {
	return &v0.Metadata{
		Count:      10,
		NextCursor: "sample-cursor-token-12345",
	}
}

// sampleServerListResponse returns a sample server list response
func sampleServerListResponse() *v0.ServerListResponse {
	return &v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			sampleActiveServer(),
			sampleFilesystemServer(),
			sampleOCIServer(),
		},
		Metadata: sampleMetadata(),
	}
}

// validServerIDs returns example valid server ID formats
func validServerIDs() []string {
	return []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b811-9dad-11d1-80b4-00c04fd430c8",
	}
}

// invalidServerIDs returns example invalid server ID formats for error testing
func invalidServerIDs() []string {
	return []string{
		"not-a-uuid",
		"12345",
		"",
		"550e8400-e29b-41d4-a716",
		"550e8400-e29b-41d4-a716-446655440000-extra",
	}
}

// validTimestamps returns valid RFC3339 timestamps for testing
func validTimestamps() []string {
	return []string{
		"2023-01-01T00:00:00Z",
		"2023-12-31T23:59:59Z",
		"2023-06-15T12:30:45.123Z",
		"2023-06-15T12:30:45+02:00",
	}
}

// invalidTimestamps returns invalid timestamp formats for error testing
func invalidTimestamps() []string {
	return []string{
		"not-a-timestamp",
		"2023-01-01",
		"2023-13-01T00:00:00Z", // Invalid month
		"2023-01-32T00:00:00Z", // Invalid day
		"2023-01-01T25:00:00Z", // Invalid hour
	}
}

// benchmarkTestData returns data specifically for benchmark tests
func benchmarkTestData() struct {
	LargeLimit    int
	SmallLimit    int
	SearchTerms   []string
	VersionFilter string
} {
	return struct {
		LargeLimit    int
		SmallLimit    int
		SearchTerms   []string
		VersionFilter string
	}{
		LargeLimit:    100,
		SmallLimit:    5,
		SearchTerms:   []string{"test", "file", "server"},
		VersionFilter: "latest",
	}
}