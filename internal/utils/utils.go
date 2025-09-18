package utils

import (
	"strings"
)

var transportTypeMap = map[string]string{
	"stdio": "stdio",
	"http":  "streamable-http",
	"sse":   "sse",
}

var reverseTransportTypeMap = map[string]string{
	"stdio":           "stdio",
	"streamable-http": "http",
	"sse":             "sse",
}

func MapToRegistryTransportType(userTransportType string) string {
	if registryType, exists := transportTypeMap[strings.ToLower(userTransportType)]; exists {
		return registryType
	}
	return userTransportType
}

func MapFromRegistryTransportType(registryTransportType string) string {
	if userType, exists := reverseTransportTypeMap[strings.ToLower(registryTransportType)]; exists {
		return userType
	}
	return registryTransportType
}

func NormalizeAndDeduplicateStrings(input []string) []string {
	if len(input) == 0 {
		return []string{}
	}

	var result []string
	seen := make(map[string]bool)

	for _, item := range input {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}

		if !seen[normalized] {
			result = append(result, normalized)
			seen[normalized] = true
		}
	}

	return result
}
