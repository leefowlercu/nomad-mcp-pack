# nomad-mcp-pack

A command-line utility that generates HashiCorp Nomad Pack definitions from MCP (Model Context Protocol) servers registered in the official MCP Registry.

## Overview

`nomad-mcp-pack` bridges the gap between the MCP ecosystem and HashiCorp Nomad by automatically generating Nomad Pack templates from MCP server definitions. It interfaces with the official MCP Server Registry at [registry.modelcontextprotocol.io](https://registry.modelcontextprotocol.io) to fetch server metadata and create deployment-ready Nomad Packs.

## Features

### ✅ Implemented Features

- **MCP Registry Client**: Full HTTP client for MCP Registry API v1.0.0 with retry logic, pagination, and semantic versioning
- **Pack Generation System**: Complete Nomad Pack generation from MCP servers with dual output modes (directory/archive)
- **Template System**: Go text/template based rendering with embedded templates for all supported package types
- **Package Type Support**: NPM, PyPI, OCI (Docker), and NuGet with specific job templates
- **Watch Mode**: Continuous monitoring with state management, concurrent generation, and graceful shutdown
- **Configuration System**: Hierarchical configuration using Viper (files, environment variables, defaults)
- **Integration Testing**: Comprehensive test suite with local registry support via git submodule
- **CLI Framework**: Cobra-based command structure with generate, watch, and server commands
- **Server Specification Parsing**: Robust parsing and validation of MCP server specifications

### 🚧 Planned Features

- **HTTP API**: Server mode provides REST API endpoints for programmatic access
- **Watch Mode Terminal UI Option**: Create a Terminal UI for operating in watch mode 

### 🆕 Recent Improvements

- **Watch Mode Implementation**: Complete watch functionality with state management, concurrent generation, and graceful shutdown
- **Consistent Logging**: All logs now go to stderr across all environments for reliable redirection
- **Graceful Shutdown**: Custom error handling ensures Ctrl-C exits cleanly without error messages
- **Enhanced Test Coverage**: Multiple packages now have 100% test coverage (genutils, serversearchutils, watchutils)
- **Robust Error Handling**: Improved error types and handling throughout the application

### To-Do Refactoring:

- **Factor Registry Package into SDK**: Use the code in the `registry` package to create an MCP Registry Go SDK

### Project Structure

```
nomad-mcp-pack/
├── cmd/                                  # Package `cmd` - Command line interface
│   ├── generate/                         # Package `cmdgenerate` - Generate command implementation
│   │   └── generate.go                   # Generate command implementation with full functionality
│   ├── server/                           # Package `cmdserver` - HTTP server command
│   │   └── server.go                     # Server command (to be implemented)
│   ├── watch/                            # Package `cmdwatch` - Watch command for continuous monitoring
│   │   └── watch.go                      # Watch command with complete implementation
│   └── nomadmcppack.go                   # Root command and CLI setup
├── demo/                                 # Demo scripts and documentation
│   ├── demo-generate.sh                  # Interactive demo script for generate command
│   └── README-DEMO.md                    # Demo documentation and usage
├── pkg/                                  # Package `pkg` - Public packages
│   ├── generate/                         # Package `generate` - Pack generation functionality
│   │   ├── generate.go                   # Core pack generation logic with dual output modes
│   │   ├── generate_test.go              # Comprehensive unit tests for generation logic
│   │   ├── templates.go                  # Template rendering system with embedded templates
│   │   ├── templates_test.go             # Unit tests for template rendering
│   │   ├── archive.go                    # ZIP archive creation functionality
│   │   ├── archive_test.go               # Unit tests for archive creation
│   │   └── templates/                    # Embedded Nomad Pack templates
│   │       ├── metadata.hcl.tmpl         # Pack metadata template
│   │       ├── outputs.tpl.tmpl          # Pack outputs template
│   │       ├── readme.md.tmpl            # Pack documentation template
│   │       ├── job-oci.nomad.tmpl        # Docker/OCI job template
│   │       ├── job-npm.nomad.tmpl        # NPM package job template
│   │       ├── job-pypi.nomad.tmpl       # PyPI package job template
│   │       └── job-nuget.nomad.tmpl      # NuGet package job template
│   ├── registry/                         # Package `registry` - MCP Registry client
│   │   ├── registry.go                   # HTTP client for MCP Registry API v1.0.0
│   │   └── registry_test.go              # Comprehensive unit tests for registry client
│   └── server/                           # Package `server` - Server utilities (placeholder)
├── internal/                             # Package `internal` - Private application code
│   ├── config/                           # Package `config` - Configuration management
│   │   └── config.go                     # Viper-based configuration with env/file/defaults
│   ├── ctxutils/                         # Package `ctxutils` - Context utilities
│   │   └── ctxutils.go                   # Request-scoped logging, tracing, etc. (to be implemented)
│   ├── genutils/                         # Package `genutils` - Generate command utilities
│   │   ├── genutils.go                   # Server specification parsing and validation
│   │   └── genutils_test.go              # Comprehensive unit tests for server spec parsing (100% coverage)
│   ├── logutils/                         # Package `logutils` - Logging utilities
│   │   └── logutils.go                   # Logger initialization with consistent stderr output
│   ├── serversearchutils/                # Package `serversearchutils` - Server search utilities
│   │   ├── serversearchutils.go          # Server name validation and search specs
│   │   └── serversearchutils_test.go     # Comprehensive unit tests (100% coverage)
│   ├── watch/                            # Package `watch` - Watch functionality
│   │   ├── watcher.go                    # Main watcher with polling, filtering, and concurrent generation
│   │   ├── state.go                      # State management with atomic persistence
│   │   ├── errors.go                     # Custom error types for graceful shutdown
│   │   ├── watcher_test.go               # Comprehensive unit tests (88.1% coverage)
│   │   └── state_test.go                 # State management tests with table-driven patterns
│   └── watchutils/                       # Package `watchutils` - Watch command utilities
│       ├── watchutils.go                 # Filter parsing, validation, and watch configuration
│       └── watchutils_test.go            # Complete unit tests (100% coverage)
├── tests/                                # Test suite
│   └── integration/                      # Integration tests
│       ├── registry/                     # Git submodule: MCP Registry (v1.0.0)
│       ├── registry_integration_test.go  # Integration tests against local registry
│       ├── helpers_test.go               # Test utilities and registry detection
│       ├── fixtures_test.go              # Test data and sample server definitions
│       └── README.md                     # Integration testing documentation
├── reference/                            # Reference documentation
│   └── registry-api-v1.0.0.yaml          # OpenAPI v1.0.0 specification for MCP Registry
├── CLAUDE.md                             # Agent instructions for AI coding assistants
├── config.yaml.example                   # Example configuration file
├── main.go                               # Application entry point
├── go.mod                                # Go module definition
├── go.sum                                # Go dependency checksums
├── Makefile                              # Build automation with registry targets
└── README.md                             # Project documentation
```

## Installation

### From Source

Requires Go 1.25.1 or later:

```bash
git clone https://github.com/leefowlercu/nomad-mcp-pack.git
cd nomad-mcp-pack
make install
```

This will build the binary and install it to `~/go/bin/`.

## Usage

### ✅ Generate Command (Implemented)

The `generate` command creates Nomad Packs from MCP servers registered in the official registry:

```bash
# Generate a pack for the latest version of an MCP server
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest

# Generate a pack for a specific version
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@0.0.1-seed

# Generate to a specific output directory
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --output-dir ./my-packs

# Generate as ZIP archive instead of directory
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --output-type archive

# Preview what would be generated without creating files
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --dry-run

# Force overwrite existing pack
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --force

# Specify package type (default 'oci' (Docker))
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --package-type npm

# Allow deprecated servers
nomad-mcp-pack generate deprecated-server/example@latest --allow-deprecated
```

### ✅ Watch Mode (Implemented)

Continuously monitor the registry for new or updated servers:

```bash
# Watch with default settings (300 second intervals)
nomad-mcp-pack watch

# Custom output directory and interval
nomad-mcp-pack watch --output-dir ./packs --poll-interval 600

# Watch with filtering by package type (only supported types)
nomad-mcp-pack watch --filter-package-type npm,oci

# Watch specific servers by name (comma-separated)
nomad-mcp-pack watch --filter-names "io.github.datastax/astra-db-mcp,io.github.ruvnet/claude-flow"

# Generate as ZIP archives instead of directories
nomad-mcp-pack watch --output-type archive

# Control concurrency (default: 5)
nomad-mcp-pack watch --max-concurrent 10

# Custom state file location
nomad-mcp-pack watch --state-file ./my-watch-state.json

# Dry run to see what would be generated
nomad-mcp-pack watch --dry-run

# Capture logs to file (all logs go to stderr)
nomad-mcp-pack watch 2>watch.log

# Stop gracefully with Ctrl-C (no error messages)
```

### 🚧 HTTP Server Mode (Planned)

Run as an HTTP server for programmatic access:

```bash
# Start on default port 8080
nomad-mcp-pack server

# Custom address
nomad-mcp-pack server --addr :9090

# With custom timeouts
nomad-mcp-pack server --read-timeout 30 --write-timeout 30
```

### Global Options

- `--mcp-registry-url`: Override the default MCP Registry URL (default: `https://registry.modelcontextprotocol.io`)
- `--help`: Show help for any command
- `--version`: Display version information

### Generated Pack Structure

Each generated pack contains:
- **metadata.hcl**: Pack metadata, name, description, and version
- **variables.hcl**: Configurable variables for environment and arguments
- **outputs.tpl**: Template for pack outputs 
- **README.md**: Generated documentation with usage instructions
- **templates/mcp-server.nomad.tpl**: Nomad job template for the specific package type

### Supported Package Types

- **OCI (Docker)** - Default: Uses Docker driver for containerized deployment
- **NPM** - Uses exec driver with `npx` to run Node.js packages
- **PyPI** - Uses exec driver with `pip` to install and run Python packages  
- **NuGet** - Uses exec driver with `dotnet tool` for .NET packages

## Current Implementation Status

### ✅ Completed Components

#### Pack Generation System (`pkg/generate/`)
- **Complete Implementation**: 588 lines of production code with 73.5% test coverage
- **Dual Output Modes**: Generate packs as directories or ZIP archives
- **Template-Based Generation**: Embedded Go templates for all pack components
- **Package Type Support**: NPM, PyPI, OCI (Docker), and NuGet with specific job templates
- **Force and Dry-Run**: Optional force overwrite and preview modes
- **Comprehensive Testing**: Unit tests for generation, templates, and archive creation

#### MCP Registry Client (`pkg/registry/`)
- **Full API v1.0.0 Support**: All query parameters (cursor, limit, updated_since, search, version)  
- **Robust HTTP Client**: Retry logic with exponential backoff, proper timeout handling (74.0% test coverage)
- **Semantic Versioning**: Client-side version comparison to find latest active servers
- **Comprehensive Testing**: Unit tests with mock servers + integration tests against real registry
- **Convenience Methods**: GetLatestActiveServer, SearchServers, GetUpdatedServers, etc.

#### Generate Command (`cmd/generate/`)
- **Full CLI Implementation**: Complete generate command with all options
- **Server Specification Parsing**: Robust parsing of `server@version` specifications
- **Configuration Integration**: Uses hierarchical configuration system
- **Error Handling**: Descriptive errors for invalid servers, missing packages, etc.

#### Configuration System (`internal/config/`)
- **Hierarchical Configuration**: Command flags > Environment variables > Config files > Defaults
- **Viper Integration**: YAML config files with automatic environment variable binding
- **Validation**: Early validation with clear error messages for invalid values
- **Environment Helpers**: IsDev(), IsProd(), IsNonProd() convenience methods

#### Watch System (`internal/watch/`)
- **Complete Implementation**: 300+ lines of production code with 88.1% test coverage
- **Continuous Monitoring**: Configurable interval polling with UpdatedSince filtering for efficiency
- **State Management**: Atomic persistence using JSON files with mutex synchronization
- **Concurrent Generation**: Semaphore-controlled concurrent pack generation with configurable limits
- **Server Filtering**: Support for filtering by server names and package types
- **Graceful Shutdown**: Custom error types for clean Ctrl-C termination without error messages
- **Comprehensive Testing**: HTTP test servers, table-driven tests, context control for infinite loops
- **Output Type Support**: Configuration for pack directory vs archive output

#### Integration Testing (`tests/integration/`)
- **Local Registry**: Git submodule pinned to MCP Registry v1.0.0
- **Automatic Setup**: `make registry-up` starts local registry with Docker Compose
- **Adaptive Tests**: Work with any available registry data, skip when registry unavailable
- **Performance Testing**: Concurrent operations, timeout handling, pagination

#### Build System (`Makefile`)
- **Registry Management**: Targets for starting/stopping/updating local registry
- **Test Separation**: Separate unit and integration test targets
- **Development Workflow**: Build, install, and clean targets

### 🚧 In Progress Components

*No components currently in progress - all core functionality is complete.*

### ⏳ Planned Components

#### Watch Functionality (`internal/watch/`)
- **State Management**: JSON-based state file to track generated packs
- **Polling Logic**: Configurable interval polling of registry for updates
- **Filtering**: Package type and server name filtering
- **TUI Option**: Terminal UI for interactive watch mode

#### Server Mode (`cmd/server/`)
- **HTTP API**: REST endpoints for pack generation
- **Registry Health Checks**: Registry connectivity and system health endpoints
- **Server Health Checks**: Server health endpoints

#### Template System (*Completed - moved to `pkg/generate/templates/`*)
- **Nomad Pack Templates**: Go templates for generating pack structure
- **Package Type Support**: NPM, PyPI, OCI (Docker), NuGet specific logic

**Note**: Core pack generation functionality is now complete and ready for use. The generate command is fully functional for creating Nomad Packs from MCP servers.

### Building Locally

```bash
make build
```

The binary will be created in the current directory as `nomad-mcp-pack`.

## Development Setup

### Initialize Submodules

This project uses the MCP Registry as a submodule for integration testing, located at `tests/integration/registry/`. When cloning the repository, use the `--recursive` flag to automatically initialize submodules:

```bash
git clone --recursive https://github.com/leefowlercu/nomad-mcp-pack.git
```

If you've already cloned the repository, initialize the submodules manually:

```bash
git submodule update --init --recursive
```

### Integration Testing

The project includes comprehensive integration tests that run against a real MCP Registry instance. The registry is included as a Git submodule at `tests/integration/registry/` pinned to v1.0.0 for consistent testing.

#### Features
- **Real Registry Testing**: Tests against actual MCP Registry API v1.0.0
- **Automatic Setup**: Local registry with `make registry-up` using Docker Compose
- **Adaptive Tests**: Work with any available registry data, skip when unavailable  
- **Performance Validation**: Concurrent operations, pagination, timeout handling
- **Version Pinning**: Registry submodule ensures consistent test environment

#### Quick Start

```bash
# Initialize submodule and start registry
make registry-up

# Verify registry is running  
curl http://localhost:8080/v0/health

# Run integration tests
make test-integration

# Stop registry when done
make registry-down
```

#### Test Commands

```bash
# Unit tests (fast, no external dependencies)
make test-unit

# Integration tests (requires local registry)
make test-integration
make test-integration-verbose    # With detailed output

# Run tests with different configurations
INTEGRATION_TEST_REGISTRY_URL=http://custom:8080 make test-integration
SKIP_INTEGRATION_TESTS=true go test ./...
go test -short ./...            # Skip integration tests
```

#### Registry Management

```bash
make registry-up                # Start local MCP Registry with Docker Compose
make registry-down              # Stop local MCP Registry
make registry-logs              # View registry container logs  
make registry-update            # Update registry submodule to latest
make registry-init              # Initialize registry submodule only
```

#### Integration Test Behavior
- **Automatic Skipping**: Skip when registry unavailable or `SKIP_INTEGRATION_TESTS=true`
- **Read-Only**: Tests don't modify registry data
- **Data Adaptive**: Work gracefully with minimal or empty registry data
- **Error Resilient**: Handle network issues, timeouts, and API changes

## Running Demonstrations

The project includes interactive demonstration scripts that showcase the functionality of `nomad-mcp-pack` commands. These demos are designed for presentations, training, and testing.

### Available Demos

#### Generate Command Demo
Located in `demo/demo-generate.sh`, this comprehensive demo showcases all features of the `generate` command including version resolution, output types, package types, and error handling.

**Quick Start:**
```bash
# Interactive mode (recommended for presentations)
./demo/demo-generate.sh

# Automatic mode (for testing)
./demo/demo-generate.sh auto

# Show help
./demo/demo-generate.sh help
```

For detailed information about the demo, see [demo/README-DEMO.md](demo/README-DEMO.md).

### Demo Features
- **Interactive presentations** with pause points for command explanations
- **Clean output** with suppressed logging for presentation clarity
- **Automatic cleanup** of all generated artifacts
- **Multiple demo servers** to showcase different package types
- **Error handling** demonstrations

Additional demos for `watch` and `server` commands will be added as those features are implemented.

## Generated Pack Structure

Each generated Nomad Pack includes:

```
<server-id>/
├── metadata.hcl              # Pack metadata and description
├── variables.hcl             # Input variables with validation
├── outputs.tpl               # Output template
├── templates/
│   └── mcp-server.nomad.tpl  # Main Nomad job template
└── README.md                 # Pack documentation
```

## Development

### Prerequisites

- Go 1.25.1 or later
- Make
- HashiCorp Nomad (for testing generated packs)

### Building

```bash
# Build the binary
make build

# Build with custom version
make build VERSION=1.0.0

# Run tests
make test

# Clean and rebuild
make rebuild
```

## Configuration

Configuration can be provided through multiple sources with the following precedence (highest to lowest):

1. **Command-line flags** - Override all other sources
2. **Environment variables** - Override config file and defaults
3. **Configuration file** - Override defaults
4. **Built-in defaults**

### Configuration File

The application looks for a `config.yaml` file in the following locations:
- Current directory (`./config.yaml`)
- User's home directory (`~/.nomad-mcp-pack/config.yaml`)

See `config.yaml.example` for a complete example configuration file.

### Environment Variables

All configuration options can be set via environment variables with the prefix `NOMAD_MCP_PACK_`. Dots in configuration keys are replaced with underscores.

#### General Configuration
- `NOMAD_MCP_PACK_MCP_REGISTRY_URL`: Override the default MCP Registry URL (defaults to https://registry.modelcontextprotocol.io)
- `NOMAD_MCP_PACK_OUTPUT_DIR`: Default output directory for generated packs (defaults to ./packs)
- `NOMAD_MCP_PACK_OUTPUT_TYPE`: Output type for generated packs (packdir or archive) - defaults to packdir
- `NOMAD_MCP_PACK_LOG_LEVEL`: Set logging level (debug, info, warn, error) - defaults to info
- `NOMAD_MCP_PACK_ENV`: Environment mode (dev, nonprod, prod) - affects log output format, defaults to prod

#### Server Configuration
- `NOMAD_MCP_PACK_SERVER_ADDR`: Server address to bind to (defaults to :8080)
- `NOMAD_MCP_PACK_SERVER_READ_TIMEOUT`: Server read timeout in seconds (defaults to 10)
- `NOMAD_MCP_PACK_SERVER_WRITE_TIMEOUT`: Server write timeout in seconds (defaults to 10)

#### Watch Configuration
- `NOMAD_MCP_PACK_WATCH_POLL_INTERVAL`: Poll interval in seconds for watch mode (defaults to 300)
- `NOMAD_MCP_PACK_WATCH_FILTER_NAMES`: Comma-separated list of server names to filter (empty = all servers)
- `NOMAD_MCP_PACK_WATCH_FILTER_PACKAGE_TYPES`: Comma-separated list of package types to filter (defaults to npm,pypi,oci,nuget)
- `NOMAD_MCP_PACK_WATCH_STATE_FILE`: State file to persist watch state between runs (defaults to watch.json)
- `NOMAD_MCP_PACK_WATCH_MAX_CONCURRENT`: Maximum concurrent pack generations (defaults to 5)
- `NOMAD_MCP_PACK_WATCH_ENABLE_TUI`: Enable TUI mode for interactive terminal interface (defaults to false)

## Limitations

### Pack Generation Limitations
The `generate` command **only** works with MCP servers that have packages in the following registries:
- NPM (Node Package Manager)
- PyPI (Python Package Index)
- OCI (Docker/Container registries)
- NuGet (.NET packages)

The following server types **cannot** be used with the `generate` command:
- Servers with only remote hosted instances (no package definitions)
- Servers with MCPB package type

To explore available MCP servers, visit the [MCP Registry](https://registry.modelcontextprotocol.io) directly.

### Other Limitations
- Requires packages to be publicly accessible

## Related Projects

- [MCP Registry](https://registry.modelcontextprotocol.io) - Official MCP Server Registry
- [HashiCorp Nomad](https://www.nomadproject.io) - Simple and flexible workload orchestrator
- [Nomad Pack](https://github.com/hashicorp/nomad-pack) - Package manager for Nomad
- [Model Context Protocol](https://modelcontextprotocol.io) - Open protocol for AI model context
- [Server.json Docs](https://github.com/modelcontextprotocol/registry/tree/main/docs/reference/server-json) - Documentation for `server.json` MCP Server specification
- [MCP Server Schema](https://github.com/modelcontextprotocol/registry/blob/main/docs/reference/server-json/server.schema.json) - Schema for `server.json` documents
- [Official Registry API Spec](https://registry.modelcontextprotocol.io/openapi.yaml) - Official MCP Registry API specification
