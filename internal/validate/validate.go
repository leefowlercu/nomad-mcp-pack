package validate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/server"
)

func PackageType(packageType string) error {
	if packageType == "" {
		return fmt.Errorf("invalid package type format; package type must not be empty")
	}

	packageTypeLower := strings.ToLower(packageType)
	if !slices.Contains(config.ValidPackageTypes, packageTypeLower) {
		return fmt.Errorf("invalid package type %q; must be one of %v", packageTypeLower, config.ValidPackageTypes)
	}

	return nil
}

func PackageTypes(types []string, requireAtLeastOne bool) error {
	if requireAtLeastOne && len(types) == 0 {
		return fmt.Errorf("at least one package type must be specified")
	}

	validCount := 0
	for _, packageType := range types {
		packageType = strings.ToLower(strings.TrimSpace(packageType))
		if packageType == "" {
			continue // Skip empty entries
		}

		if !slices.Contains(config.ValidPackageTypes, packageType) {
			return fmt.Errorf("invalid package type %q; must be one of %v", packageType, config.ValidPackageTypes)
		}
		validCount++
	}

	if requireAtLeastOne && validCount == 0 {
		return fmt.Errorf("at least one package type must be specified")
	}

	return nil
}

func TransportType(transportType string) error {
	if transportType == "" {
		return fmt.Errorf("invalid transport type format; transport type must not be empty")
	}

	transportTypeLower := strings.ToLower(transportType)
	if !slices.Contains(config.ValidTransportTypes, transportTypeLower) {
		return fmt.Errorf("invalid transport type %q; must be one of %v", transportTypeLower, config.ValidTransportTypes)
	}

	return nil
}

func TransportTypes(types []string, requireAtLeastOne bool) error {
	if requireAtLeastOne && len(types) == 0 {
		return fmt.Errorf("at least one transport type must be specified")
	}

	validCount := 0
	for _, transportType := range types {
		transportType = strings.ToLower(strings.TrimSpace(transportType))
		if transportType == "" {
			continue // Skip empty entries
		}

		if !slices.Contains(config.ValidTransportTypes, transportType) {
			return fmt.Errorf("invalid transport type %q; must be one of %v", transportType, config.ValidTransportTypes)
		}
		validCount++
	}

	if requireAtLeastOne && validCount == 0 {
		return fmt.Errorf("at least one transport type must be specified")
	}

	return nil
}

func OutputDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("invalid output directory format; output directory must not be empty")
	}

	return nil
}

func OutputType(outputType string) error {
	if outputType == "" {
		return fmt.Errorf("invalid output type format; output type must not be empty")
	}

	outputTypeLower := strings.ToLower(outputType)
	if !slices.Contains(config.ValidOutputTypes, outputTypeLower) {
		return fmt.Errorf("invalid output type %q; must be one of %v", outputTypeLower, config.ValidOutputTypes)
	}

	return nil
}

func ServerNames(names []string) error {
	if len(names) == 0 {
		return nil // Empty filter is valid
	}

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue // Skip empty entries
		}

		// Validate server name using shared validation function
		if _, err := server.ParseNameSpec(name); err != nil {
			return err
		}
	}

	return nil
}

func PollInterval(interval int) error {
	if interval < config.MinPollInterval {
		return fmt.Errorf("poll interval must be at least %d seconds, got %d", config.MinPollInterval, interval)
	}
	return nil
}

func StateFile(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("state file path cannot be empty")
	}
	return nil
}

func MaxConcurrent(max int) error {
	if max < config.MinMaxConcurrent {
		return fmt.Errorf("max concurrent must be at least %d, got %d", config.MinMaxConcurrent, max)
	}
	return nil
}
