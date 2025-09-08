package genutils

import (
	"fmt"
	"strings"
)

type ServerSpec struct {
	ServerName string
	Version    string
}

// ParseServerSpec parses a server specification in the format <mcp-server@version>
// where mcp-server is the server name (e.g., "io.github.datastax/astra-db-mcp")
// and version is either a semver string (e.g., "0.0.1-seed") or "latest"
// additionally the server name portion must consist of a namespace and name separated by a "/"
func ParseServerSpec(spec string) (*ServerSpec, error) {
	if spec == "" {
		return nil, fmt.Errorf("server specification cannot be empty")
	}

	// Split by @ to separate server name from version
	parts := strings.Split(spec, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid server specification format: expected <mcp-server@version>, got %q", spec)
	}

	serverName := strings.TrimSpace(parts[0])
	version := strings.TrimSpace(parts[1])

	if serverName == "" {
		return nil, fmt.Errorf("server name cannot be empty in specification %q", spec)
	}

	if version == "" {
		return nil, fmt.Errorf("version cannot be empty in specification %q", spec)
	}

	// Validate server name format (should contain namespace/name pattern)
	if !strings.Contains(serverName, "/") {
		return nil, fmt.Errorf("invalid server name format %q: expected format like 'example.com/server'", serverName)
	}

	// Basic validation that server name has reasonable structure
	nameParts := strings.Split(serverName, "/")
	if len(nameParts) != 2 {
		return nil, fmt.Errorf("invalid server name format %q: expected exactly one '/' separator", serverName)
	}

	if nameParts[0] == "" || nameParts[1] == "" {
		return nil, fmt.Errorf("invalid server name format %q: namespace and name parts cannot be empty", serverName)
	}

	return &ServerSpec{
		ServerName: serverName,
		Version:    version,
	}, nil
}

func (s *ServerSpec) IsLatest() bool {
	return strings.ToLower(s.Version) == "latest"
}

func (s *ServerSpec) String() string {
	return fmt.Sprintf("%s@%s", s.ServerName, s.Version)
}
