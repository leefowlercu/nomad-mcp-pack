package output

import (
	"fmt"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
)

// isSilent checks if the silent flag is enabled by reading the current configuration
func isSilent() bool {
	cfg, err := config.GetConfig()
	if err != nil {
		// If we can't get config, default to not silent to ensure output is shown
		return false
	}
	return cfg.Silent
}

// Print outputs a message to stdout if not in silent mode
func Print(format string, args ...any) {
	if !isSilent() {
		fmt.Printf(format, args...)
	}
}

// Println outputs a message with newline to stdout if not in silent mode
func Println(format string, args ...any) {
	if !isSilent() {
		fmt.Printf(format+"\n", args...)
	}
}

// Success outputs a success message with checkmark to stdout if not in silent mode
func Success(format string, args ...any) {
	if !isSilent() {
		fmt.Printf("✓ "+format+"\n", args...)
	}
}

// Info outputs an informational message to stdout if not in silent mode
func Info(format string, args ...any) {
	if !isSilent() {
		fmt.Printf(format+"\n", args...)
	}
}

// Warning outputs a warning message to stdout if not in silent mode
func Warning(format string, args ...any) {
	if !isSilent() {
		fmt.Printf("Warning: "+format+"\n", args...)
	}
}

// Failure outputs a failure message with X mark to stdout if not in silent mode
func Failure(format string, args ...any) {
	if !isSilent() {
		fmt.Printf("✗ "+format+"\n", args...)
	}
}

// Progress outputs a progress message to stdout if not in silent mode
func Progress(format string, args ...any) {
	if !isSilent() {
		fmt.Printf("⏳ "+format+"\n", args...)
	}
}
