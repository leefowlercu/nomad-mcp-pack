package generator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/registry/pkg/model"
)

// NPMPackageInfo represents metadata from the NPM registry
type NPMPackageInfo struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Bin     any               `json:"bin"` // Can be string or map[string]string
	Main    string            `json:"main"`
	Scripts map[string]string `json:"scripts"`
}

// ExecutionPattern represents how an NPM package should be executed
type ExecutionPattern string

const (
	ExecutionPatternGlobalBin  ExecutionPattern = "global-bin"  // npm install -g && command
	ExecutionPatternNodeDirect ExecutionPattern = "node-direct" // npm install && node script
	ExecutionPatternNPX        ExecutionPattern = "npx"         // npx package@version
)

// NPMExecutionData contains resolved execution information for templates
type NPMExecutionData struct {
	Pattern    ExecutionPattern
	BinCommand string // The command to run (for global-bin pattern)
	ScriptPath string // The script path (for node-direct pattern)
}

// fetchNPMPackageInfo retrieves package metadata from the NPM registry
func fetchNPMPackageInfo(packageID, version string) (*NPMPackageInfo, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/%s", packageID, version)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NPM package info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NPM registry returned status %d", resp.StatusCode)
	}

	var info NPMPackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode NPM package info: %w", err)
	}

	return &info, nil
}

// parseBinField extracts the bin command from various bin field formats
func parseBinField(bin any, packageName string) (string, bool) {
	if bin == nil {
		return "", false
	}

	switch v := bin.(type) {
	case string:
		// Simple bin: "bin": "script.js" -> use package name as command
		return packageName, true
	case map[string]any:
		// Object bin: "bin": {"cmd": "script.js"}
		if len(v) > 0 {
			// Return first command name
			for cmdName := range v {
				return cmdName, true
			}
		}
	}

	return "", false
}

// resolveNPMExecutionPattern determines how to execute an NPM package
func resolveNPMExecutionPattern(pkg *model.Package) (*NPMExecutionData, error) {
	data := &NPMExecutionData{}

	// Try to fetch package info from NPM registry
	npmInfo, err := fetchNPMPackageInfo(pkg.Identifier, pkg.Version)

	if err == nil && npmInfo != nil {
		// Check if package has bin entry
		if binCmd, hasBin := parseBinField(npmInfo.Bin, pkg.Identifier); hasBin {
			data.Pattern = ExecutionPatternGlobalBin
			data.BinCommand = binCmd

			// Special case: if there's a positional argument with .js file,
			// it might be misconfigured (like git-mcp-server)
			// In this case, ignore the positional arg and use the bin command
			return data, nil
		}

		// No bin entry - need node execution
		data.Pattern = ExecutionPatternNodeDirect

		// Check for script path in positional arguments
		for _, arg := range pkg.PackageArguments {
			if arg.Type == "positional" && arg.Value != "" {
				if strings.HasSuffix(arg.Value, ".js") || strings.HasSuffix(arg.Value, ".mjs") {
					// Normalize the path by removing leading "./"
					data.ScriptPath = strings.TrimPrefix(arg.Value, "./")
					break
				}
			}
		}

		// Fallback to main field if no positional argument
		if data.ScriptPath == "" && npmInfo.Main != "" {
			// Normalize the path by removing leading "./"
			data.ScriptPath = strings.TrimPrefix(npmInfo.Main, "./")
		}

		// If still no script path, default to index.js
		if data.ScriptPath == "" {
			data.ScriptPath = "index.js"
		}
	} else {
		// NPM registry unavailable - use heuristics

		// Check if there's a positional argument with a .js file
		hasJSPositional := false
		for _, arg := range pkg.PackageArguments {
			if arg.Type == "positional" && arg.Value != "" {
				if strings.HasSuffix(arg.Value, ".js") || strings.HasSuffix(arg.Value, ".mjs") {
					data.Pattern = ExecutionPatternNodeDirect
					// Normalize the path by removing leading "./"
					data.ScriptPath = strings.TrimPrefix(arg.Value, "./")
					hasJSPositional = true
					break
				}
			}
		}

		if !hasJSPositional {
			// Assume it has a global bin command
			// Use package name without scope as command
			data.Pattern = ExecutionPatternGlobalBin
			data.BinCommand = pkg.Identifier

			// Remove scope from command name (e.g., @scope/package -> package)
			if strings.HasPrefix(data.BinCommand, "@") {
				parts := strings.Split(data.BinCommand, "/")
				if len(parts) == 2 {
					data.BinCommand = parts[1]
				}
			}
		}
	}

	return data, nil
}
