# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`nomad-mcp-pack` is a Go utility that generates HashiCorp Nomad Pack definitions from MCP (Model Context Protocol) Servers registered in the official MCP Registry at `registry.modelcontextprotocol.io`. It bridges the gap between the MCP ecosystem and Nomad-based deployment platforms.

### Core Functionality
- Interfaces with official MCP Server Registry (`registry.modelcontextprotocol.io`) to generate Nomad Packs
- Transforms [server.json](https://github.com/modelcontextprotocol/registry/tree/main/docs/server-json) specifications into Nomad Pack templates
- Support the following MCP Server Registry upstream Registry types (NPM, PyPI, OCI (Docker), NuGet)
- Enable both CLI and HTTP Server modes for flexibility
- Provide Polling capabilities and optional filtering for continuous synchronization with official MCP Server Registry 

### Current Limitations
- Does not support Remote MCP Servers (i.e. MCP Servers with no Package definitions)

## Architecture

### Key Design Decisions

1. **Dual Mode Operation**: CLI for direct usage, HTTP server for remote & programmatic integration
2. **Registry-First**: All MCP Server metadata comes from the official MCP Server Registry
3. **Package Type Abstraction**: Intelligently handles different package registries (NPM, PyPI, OCI (Docker), NuGet)
4. **Template-Based Generation**: Uses Go text/template for flexible Nomad job generation
5. **Stateful Watching**: Maintains state to avoid regenerating unchanged packs

### Component Structure

```
nomad-mcp-pack/
├── cmd/                    # Application entry points
│   ├── nomad-mcp-pack/    # Main binary (uses spf13/cobra)
│   └── cli/               # CLI commands (generate, list, watch, server, inspect)
├── pkg/                    # Core packages
│   ├── registry/          # MCP Registry client and types
│   ├── generator/         # Nomad Pack generation logic
│   ├── watcher/           # Polling and state management
│   ├── config/            # Configuration handling
│   └── api/               # HTTP server implementation
└── internal/              # Internal utilities
    └── templates/         # Embedded Nomad Pack templates
```

## Key Components

### CLI Framework (`cmd/`)
- **Library**: Uses `spf13/cobra` for command-line interface modeling
- **Root Command**: `nomad-mcp-pack` - Main entry point with global flags
- **Subcommands**:
  - `generate` - Generate a single Nomad Pack from an MCP server
  - `list` - List available MCP servers from the registry
  - `watch` - Continuously poll registry and generate packs for new/updated servers
  - `server` - Run as HTTP server for programmatic access
  - `inspect` - Examine details of a specific MCP server
- **Flag Handling**: Cobra's built-in flag parsing with persistent and local flags
- **Command Structure**:
  ```go
  // Example command setup with cobra
  var rootCmd = &cobra.Command{
      Use:   "nomad-mcp-pack",
      Short: "Generate Nomad Packs from MCP Server Registry",
  }
  ```

### Logging Strategy
- **Library**: Uses Go standard library `log/slog` for structured logging
- **Log Levels**: Debug, Info, Warn, Error
- **Structured Fields**: Context-aware logging with key-value pairs
- **Performance**: Zero-allocation logging for common paths
- **Configuration**: Log level configurable via environment variable or CLI flag
- **Example Usage**:
  ```go
  slog.Info("generating pack",
      "server_id", serverID,
      "output_dir", outputDir,
      "package_type", packageType,
  )
  
  slog.Error("failed to fetch server",
      "error", err,
      "server_id", serverID,
      "retry_count", retries,
  )
  ```
- **Context Integration**: Pass logger through context for request-scoped logging
- **Output Format**: JSON for production, text for development

### Registry Client (`pkg/registry/`)
- **Purpose**: Interface with the MCP Registry API
- **Key Types**: `ServerJSON`, `Package`, `EnvironmentVariable`
- **API Endpoints**: 
  - `GET /v0/servers` - List servers with filtering
  - `GET /v0/servers/{id}` - Get detailed server info
- **Important**: The registry returns `server.json` format, not raw MCP protocol data

### Pack Generator (`pkg/generator/`)
- **Purpose**: Transform registry data into Nomad Pack files
- **Key Responsibilities**:
  - Select appropriate package type (prioritizes Docker > GitHub > NPM > PyPI)
  - Generate HCL templates with proper variable substitution
  - Handle environment variable requirements
  - Create resource constraints and network configuration
- **Output Files**:
  - `metadata.hcl` - Pack metadata
  - `variables.hcl` - Input variables with validation
  - `templates/server.nomad.tpl` - Main job template
  - `outputs.hcl` - Output values
  - `README.md` - Documentation

### Watcher (`pkg/watcher/`)
- **Purpose**: Continuously monitor registry for changes
- **State Management**: Tracks generated packs to avoid duplicates
- **Filtering**: Supports complex filters (by category, package type, etc.)
- **Concurrency**: Configurable parallel pack generation

## Development Guidelines

### Adding New Package Types

When adding support for a new package registry type:

1. Update `pkg/registry/types.go` if new fields are needed
2. Add case in `generator.selectPrimaryPackage()` with appropriate priority
3. Implement `generate<Type>Config()` method in generator
4. Add any type-specific variables in `generateVariables()`

Example for adding Helm support:
```go
func (g *Generator) generateHelmConfig(pkg *registry.Package) string {
    return fmt.Sprintf(`config {
        image = "alpine/helm:latest"
        command = "helm"
        args = ["install", "%s", "%s"]
    }`, pkg.Name, pkg.Version)
}
```

### Working with Templates

The Nomad job templates use a two-stage templating approach:
1. Go template processing (using `{{ }}`) - happens at pack generation time
2. Nomad Pack template processing (using `[[ ]]`) - happens at deployment time

Be careful to use the correct delimiters:
- `{{ .Server.ID }}` - Resolved when generating the pack
- `[[ .tenant_id ]]` - Resolved when deploying the pack

### Error Handling Patterns

Always wrap errors with context and log appropriately:
```go
server, err := client.GetServer(ctx, serverID)
if err != nil {
    slog.Error("failed to fetch server",
        "error", err,
        "server_id", serverID,
    )
    return fmt.Errorf("failed to fetch server %s: %w", serverID, err)
}
```

For debug-level logging of intermediate steps:
```go
slog.Debug("fetching server from registry",
    "server_id", serverID,
    "endpoint", registryURL,
)
```

### Testing Approach

1. **Unit Tests**: Test individual components in isolation
   - Registry client with mock HTTP responses
   - Generator with fixture server definitions
   - Template rendering with various input combinations

2. **Integration Tests**: Test end-to-end workflows
   - Use test tag: `// +build integration`
   - Requires Docker and Nomad for validation

3. **Template Validation**: Ensure generated HCL is valid
   ```go
   func TestGeneratedPackValidation(t *testing.T) {
       // Generate pack
       // Run `nomad pack validate` on output
       // Check for errors
   }
   ```

## Common Tasks

### Adding a New CLI Command

1. Create new file in `cmd/cli/` with cobra command definition:
   ```go
   package cli
   
   import (
       "github.com/spf13/cobra"
   )
   
   var newCmd = &cobra.Command{
       Use:   "new-command",
       Short: "Description of the command",
       RunE: func(cmd *cobra.Command, args []string) error {
           // Implementation
           return nil
       },
   }
   
   func init() {
       rootCmd.AddCommand(newCmd)
       newCmd.Flags().StringP("flag", "f", "default", "flag description")
   }
   ```
2. Add to root command in init() (happens automatically with init() function)
3. Implement business logic by calling appropriate packages
4. Use cobra's built-in validation, completion, and help generation

### Extending Registry Filtering

Add new filter criteria in `pkg/registry/filter.go`:
```go
type Filter struct {
    Categories []string
    PackageTypes []string
    RequiredEnvVars []string  // New filter
}

func (f *Filter) Matches(server *ServerSummary) bool {
    // Add matching logic
}
```

### Custom Pack Templates

To support custom templates:
1. Add template override directory support in config
2. Load custom templates in generator initialization
3. Merge with default templates, preferring custom ones

## External Integrations

### MCP Registry API
- **Base URL**: `https://registry.modelcontextprotocol.io`
- **Rate Limits**: Be mindful of API rate limits (not officially documented)
- **Caching**: Consider implementing response caching for frequently accessed servers

### Nomad Integration
- **Pack Format**: Must be compatible with Nomad Pack CLI
- **Validation**: Use `nomad pack validate` to verify generated packs
- **Variables**: Follow Nomad Pack variable naming conventions

### Multi-Tenant Platform Integration
The generated packs are designed for multi-tenant environments:
- `tenant_id` - Required variable for namespace isolation
- `deployment_id` - Unique identifier for each deployment
- `user_id` - Track who deployed the instance

## Important Constraints

1. **No Direct MCP Protocol Implementation**: This tool works with registry metadata, not MCP protocol itself
2. **Package Availability**: Generated packs assume packages are publicly accessible
3. **Transport Assumptions**: Currently assumes HTTP transport for multi-tenant deployment
4. **Security Model**: Vault integration is optional but recommended for secrets

## Configuration

### Environment Variables
```bash
NOMAD_MCP_PACK_REGISTRY_URL    # Override registry URL
NOMAD_MCP_PACK_OUTPUT_DIR       # Default output directory
NOMAD_MCP_PACK_LOG_LEVEL        # Logging level (debug, info, warn, error)
```

### Config File Structure
```yaml
registry:
  url: https://registry.modelcontextprotocol.io
  timeout: 30s
  cache_ttl: 5m

generator:
  default_resources:
    cpu: 500
    memory: 256
  enable_vault: true
  enable_consul_connect: true

watcher:
  state_file: ~/.nomad-mcp-pack/state.json
  poll_interval: 5m
```

## Debugging Tips

### Verbose Output
Use `-v` flag to enable debug-level structured logging:
```bash
nomad-mcp-pack generate mcp-github -v
# Or set via environment variable
NOMAD_MCP_PACK_LOG_LEVEL=debug nomad-mcp-pack generate mcp-github
```

### Dry Run Mode
Test generation without writing files:
```bash
nomad-mcp-pack generate mcp-github --dry-run
```

### Inspecting Registry Responses
Save raw registry responses for debugging:
```bash
curl https://registry.modelcontextprotocol.io/api/v0/servers/mcp-github | jq
```

## Future Enhancements to Consider

1. **Pack Versioning**: Track pack versions separately from server versions
2. **Dependency Resolution**: Handle MCP servers that depend on other services
3. **Resource Profiling**: Suggest resource allocations based on server characteristics
4. **Pack Repository**: Push generated packs to a Nomad Pack registry
5. **Validation Suite**: Comprehensive validation including test deployments
6. **Metrics Collection**: Track generation success/failure rates
7. **WebAssembly Support**: For WASM-based MCP servers
8. **Private Registry Support**: Handle authenticated package registries

## Code Style Guidelines

1. Use standard Go formatting (`gofmt`)
2. Follow effective Go patterns
3. Keep functions focused and testable
4. Document exported types and functions
5. Use meaningful variable names (avoid single letters except for indices)
6. Handle all errors explicitly
7. Use context for cancellation and timeouts

## Development Commands

### Building and Testing
```bash
# Build the binary
make build

# Build with custom version
make build VERSION=1.0.0

# Run tests
make test

# Clean build artifacts
make clean

# Rebuild (clean + build)
make rebuild

# Install to ~/go/bin
make install
```

### Running the CLI
```bash
# Show help
./nomad-mcp-pack --help

# List available MCP servers
./nomad-mcp-pack list

# Generate a pack for a specific server
./nomad-mcp-pack generate mcp-github -o ./packs/github

# Inspect server details
./nomad-mcp-pack inspect mcp-github

# Start HTTP server mode
./nomad-mcp-pack server --port 8080

# Watch for new/updated servers
./nomad-mcp-pack watch --output ./packs

# Use custom registry
./nomad-mcp-pack --mcp-registry-addr https://custom.registry.io list
```

## Current Implementation Status

### Completed
- ✅ Project structure following Go standards
- ✅ Cobra CLI framework with all subcommands defined
- ✅ Makefile with build, test, install targets
- ✅ Structured logging setup with `log/slog`
- ✅ Global flag for MCP registry address
- ✅ Command structure: generate, inspect, list, server, watch

### TODO - Core Implementation
The following packages need to be implemented in the `internal/` directory:

1. **`internal/generate/`** - Pack generation logic
   - Fetch server from registry
   - Select appropriate package type
   - Generate Nomad Pack files
   - Handle dry-run and force options

2. **`internal/inspect/`** - Server inspection logic
   - Fetch detailed server information
   - Format output (text, json, yaml)
   - Filter by packages or environment variables

3. **`internal/list/`** - Server listing logic
   - Query registry with filters
   - Handle pagination
   - Format output as table, json, or yaml

4. **`internal/server/`** - HTTP server implementation
   - REST API endpoints for pack operations
   - CORS support
   - TLS configuration
   - Timeout handling

5. **`internal/watch/`** - Continuous monitoring logic
   - Poll registry at intervals
   - Track state to avoid duplicates
   - Concurrent pack generation
   - State file management

### TODO - Supporting Packages
The following packages need to be created in the `pkg/` directory:

1. **`pkg/registry/`** - MCP Registry client
   - Types for ServerJSON, Package, EnvironmentVariable
   - HTTP client for registry API
   - Error handling and retries

2. **`pkg/generator/`** - Pack generation engine
   - Template processing
   - Package type selection logic
   - HCL file generation

3. **`pkg/config/`** - Configuration management
   - Load from environment variables
   - Config file parsing
   - Default values

4. **`pkg/api/`** - HTTP API handlers
   - REST endpoint implementations
   - Request/response types
   - Middleware

5. **`pkg/watcher/`** - State management for watch mode
   - State persistence
   - Change detection
   - Filter application

### TODO - Templates
Create templates in `internal/templates/`:
- `metadata.hcl.tmpl`
- `variables.hcl.tmpl`
- `outputs.hcl.tmpl`
- `server.nomad.tpl.tmpl`
- `README.md.tmpl`

## Questions or Design Decisions

When working on this codebase, consider:

1. **Should we cache registry responses?** Currently no caching, but could improve performance
2. **How to handle private packages?** Need authentication mechanism for private registries
3. **Pack naming conventions?** Currently using `mcp-{server-id}`, but could be configurable
4. **Resource defaults?** Currently hardcoded, could be smarter based on server type
5. **Multi-version support?** Should we generate packs for multiple versions of same server?

## Related Documentation

- [MCP Registry API](https://registry.modelcontextprotocol.io/docs)
- [Nomad Pack Documentation](https://developer.hashicorp.com/nomad/docs/tools/pack)
- [MCP Specification](https://modelcontextprotocol.io/specification)
- [Server.json Schema](https://github.com/modelcontextprotocol/registry/tree/main/docs/server-json)
