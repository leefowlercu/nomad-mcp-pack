package serversearchutils

import (
	"fmt"
	"strings"
)

// ServerSearchSpec represents a server search specification with separated namespace and name
type ServerSearchSpec struct {
	Namespace string // e.g., "io.github.datastax"
	Name      string // e.g., "astra-db-mcp"
	Version   string // e.g., "0.0.1-seed" or "latest"
}

// FullName returns the full server name in "namespace/name" format
func (s *ServerSearchSpec) FullName() string {
	return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
}

// IsLatest returns true if the version is "latest" (case-insensitive)
func (s *ServerSearchSpec) IsLatest() bool {
	return strings.ToLower(s.Version) == "latest"
}

// String returns the full specification in "namespace/name@version" format
func (s *ServerSearchSpec) String() string {
	return fmt.Sprintf("%s@%s", s.FullName(), s.Version)
}

// ValidateServerName validates and splits a server name into namespace and name parts
// Expected format: "namespace/name" (e.g., "io.github.datastax/astra-db-mcp")
func ValidateServerName(fullName string) (namespace, name string, err error) {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return "", "", fmt.Errorf("server name cannot be empty")
	}

	// Check for exactly one "/" separator
	if !strings.Contains(fullName, "/") {
		return "", "", fmt.Errorf("invalid server name format %q: expected format like 'io.github.example/server'", fullName)
	}

	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid server name format %q: expected exactly one '/' separator", fullName)
	}

	namespace = strings.TrimSpace(parts[0])
	name = strings.TrimSpace(parts[1])

	if namespace == "" || name == "" {
		return "", "", fmt.Errorf("invalid server name format %q: namespace and name parts cannot be empty", fullName)
	}

	return namespace, name, nil
}

// ParseServerName is an alias for ValidateServerName for clarity in usage
func ParseServerName(fullName string) (namespace, name string, err error) {
	return ValidateServerName(fullName)
}

// ParseServerSearchSpec parses a server specification in the format "namespace/name@version"
// where namespace/name is the server name (e.g., "io.github.datastax/astra-db-mcp")
// and version is either a semver string (e.g., "0.0.1-seed") or "latest"
func ParseServerSearchSpec(spec string) (*ServerSearchSpec, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("server specification cannot be empty")
	}

	// Split by @ to separate server name from version
	parts := strings.Split(spec, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid server specification format: expected <namespace/name@version>, got %q", spec)
	}

	serverName := strings.TrimSpace(parts[0])
	version := strings.TrimSpace(parts[1])

	if serverName == "" {
		return nil, fmt.Errorf("server name cannot be empty in specification %q", spec)
	}

	if version == "" {
		return nil, fmt.Errorf("version cannot be empty in specification %q", spec)
	}

	// Validate and parse the server name
	namespace, name, err := ValidateServerName(serverName)
	if err != nil {
		return nil, err
	}

	return &ServerSearchSpec{
		Namespace: namespace,
		Name:      name,
		Version:   version,
	}, nil
}