package watchutils

import (
	"fmt"
	"slices"
	"strings"

	"github.com/leefowlercu/nomad-mcp-pack/internal/serversearchutils"
)

// ServerNameFilter represents a filter for MCP server names
type ServerNameFilter struct {
	Names []string
}

// PackageTypeFilter represents a filter for package types
type PackageTypeFilter struct {
	Types []string
}

// ValidPackageTypes defines the package types supported by the watch command
var ValidPackageTypes = []string{"npm", "pypi", "oci", "nuget"}

// ParseFilterNames parses a comma-separated list of server names and validates each one
// Returns an empty slice if the input is empty (no filter applied)
// Each server name must be in the format namespace/name
func ParseFilterNames(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}, nil
	}

	parts := strings.Split(input, ",")
	var names []string
	seen := make(map[string]bool)

	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue // Skip empty parts
		}

		// Validate server name using shared validation function
		if _, _, err := serversearchutils.ValidateServerName(name); err != nil {
			return nil, err
		}

		// Add to result if not already seen (deduplication)
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}

	return names, nil
}

// ParsePackageTypes parses a comma-separated list of package types and validates each one
// At least one valid package type must be specified
// Valid types are: npm, pypi, oci, nuget (case-insensitive)
func ParsePackageTypes(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("at least one package type must be specified")
	}

	parts := strings.Split(input, ",")
	var types []string
	seen := make(map[string]bool)

	for _, part := range parts {
		packageType := strings.ToLower(strings.TrimSpace(part))
		if packageType == "" {
			continue // Skip empty parts
		}

		// Validate package type
		if !isValidPackageType(packageType) {
			return nil, fmt.Errorf("invalid package type %q: must be one of %v", packageType, ValidPackageTypes)
		}

		// Add to result if not already seen (deduplication)
		if !seen[packageType] {
			types = append(types, packageType)
			seen[packageType] = true
		}
	}

	if len(types) == 0 {
		return nil, fmt.Errorf("at least one package type must be specified")
	}

	return types, nil
}

// ValidateWatchConfig validates watch configuration parameters
func ValidateWatchConfig(pollInterval int, stateFile string, maxConcurrent int) error {
	// Validate poll interval
	if pollInterval < 30 {
		return fmt.Errorf("poll interval must be at least 30 seconds, got %d", pollInterval)
	}

	// Validate state file
	stateFile = strings.TrimSpace(stateFile)
	if stateFile == "" {
		return fmt.Errorf("state file path cannot be empty")
	}

	// Validate max concurrent
	if maxConcurrent < 1 {
		return fmt.Errorf("max concurrent must be at least 1, got %d", maxConcurrent)
	}

	return nil
}

// Matches returns true if the server name matches the filter
// If the filter is empty, all server names match
func (f *ServerNameFilter) Matches(serverName string) bool {
	if len(f.Names) == 0 {
		return true // Empty filter matches all
	}

	return slices.Contains(f.Names, serverName)
}

// Matches returns true if the package type matches the filter (case-insensitive)
func (f *PackageTypeFilter) Matches(packageType string) bool {
	packageTypeLower := strings.ToLower(packageType)

	for _, filterType := range f.Types {
		if strings.ToLower(filterType) == packageTypeLower {
			return true
		}
	}
	return false
}

// isValidPackageType checks if a package type is in the list of valid types
func isValidPackageType(packageType string) bool {
	packageTypeLower := strings.ToLower(packageType)
	return slices.Contains(ValidPackageTypes, packageTypeLower)
}