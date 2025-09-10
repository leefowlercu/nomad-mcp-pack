# nomad-mcp-pack

A command-line utility that generates HashiCorp Nomad Pack definitions from MCP (Model Context Protocol) servers registered in the official MCP Registry.

## Overview

`nomad-mcp-pack` bridges the gap between the MCP ecosystem and HashiCorp Nomad by automatically generating Nomad Pack templates from MCP server definitions. It interfaces with the official MCP Server Registry at [registry.modelcontextprotocol.io](https://registry.modelcontextprotocol.io) to fetch server metadata and create deployment-ready Nomad Packs.

## Features

### ✅ Implemented Features

- **MCP Registry Client**: Full HTTP client for MCP Registry API v1.0.0 with retry logic, pagination, and semantic versioning
- **Configuration System**: Hierarchical configuration using Viper (files, environment variables, defaults)
- **Integration Testing**: Comprehensive test suite with local registry support via git submodule
- **CLI Framework**: Cobra-based command structure with generate, watch, and server commands

### 🚧 Planned Features

- **Automatic Pack Generation**: Generate Nomad Packs from MCP servers with supported package types (NPM, PyPI, OCI/Docker, NuGet)
- **Version Management**: Support for specific versions or latest version selection
- **Continuous Monitoring**: Watch mode for automatic pack generation when new servers are added or updated
- **HTTP API**: Server mode provides REST API endpoints for programmatic access
- **State Management**: Track generated packs to avoid duplicates in watch mode
- **Flexible Filtering**: Filter servers by package type and names in watch mode
- **Watch Mode Terminal UI Option**: Create a Terminal UI for operating in watch mode 

### To-Do Refactoring:

- **Factor Registry Package into SDK**: Use the code in the `registry` package to create an MCP Registry Go SDK

## Installation

### From Source

Requires Go 1.25.1 or later:

```bash
git clone https://github.com/leefowlercu/nomad-mcp-pack.git
cd nomad-mcp-pack
make install
```

This will build the binary and install it to `~/go/bin/`.

## Current Implementation Status

### ✅ Completed Components

#### MCP Registry Client (`pkg/registry/`)
- **Full API v1.0.0 Support**: All query parameters (cursor, limit, updated_since, search, version)  
- **Robust HTTP Client**: Retry logic with exponential backoff, proper timeout handling
- **Semantic Versioning**: Client-side version comparison to find latest active servers
- **Comprehensive Testing**: Unit tests with mock servers + integration tests against real registry
- **Convenience Methods**: GetLatestActiveServer, SearchServers, GetUpdatedServers, etc.

#### Configuration System (`internal/config/`)
- **Hierarchical Configuration**: Command flags > Environment variables > Config files > Defaults
- **Viper Integration**: YAML config files with automatic environment variable binding
- **Validation**: Early validation with clear error messages for invalid values
- **Environment Helpers**: IsDev(), IsProd(), IsNonProd() convenience methods

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

#### CLI Commands (`cmd/`)
- **Structure Complete**: Cobra-based generate, watch, and server commands defined
- **Integration Pending**: Commands need implementation using registry client and config system

#### Pack Generation (`pkg/generate/`)
- **Design Complete**: Template-based approach using Go text/template
- **Implementation Needed**: Logic to transform MCP server.json to Nomad Pack structure

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

#### Template System (`internal/templates/`)
- **Nomad Pack Templates**: Go templates for generating pack structure
- **Package Type Support**: NPM, PyPI, OCI (Docker), NuGet specific logic

**Note**: Core functionality is actively under development. The MCP Registry client and configuration system are production-ready, but pack generation features are not yet implemented.

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

## Usage

> **⚠️ Important**: The commands below show the intended usage. Core functionality is still under development and these commands will not work until the implementation is complete.

### Basic Commands

#### Generate a Nomad Pack

```bash
# Generate pack for the latest version of a server
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest

# Generate pack for a specific version
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@0.0.1-seed

# Specify output directory
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --output-dir ./packs/astra-db

# Dry run to preview without creating files
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --dry-run

# Force overwrite existing pack
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@0.0.1-seed --force

# Specify preferred package type when multiple are available
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --package-type oci
```

#### Watch Mode

Continuously monitor the registry for new or updated servers:

```bash
# Watch with default settings
nomad-mcp-pack watch

# Custom output directory and interval
nomad-mcp-pack watch --output-dir ./packs --poll-interval 300

# Watch with filtering by package type (only supported types)
nomad-mcp-pack watch --filter-package-type npm,oci

# Watch specific servers by name
nomad-mcp-pack watch --filter-names io.github.datastax/astra-db-mcp

# Enable Terminal UI mode
nomad-mcp-pack watch --enable-tui

# Dry run to see what would be generated
nomad-mcp-pack watch --dry-run
```

#### HTTP Server Mode

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

## Generated Pack Structure

Each generated Nomad Pack includes:

```
<server-id>/
├── metadata.hcl           # Pack metadata and description
├── variables.hcl          # Input variables with validation
├── outputs.hcl            # Output values
├── templates/
│   └── server.nomad.tpl   # Main Nomad job template
└── README.md              # Pack documentation
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

### Project Structure

```
nomad-mcp-pack/
├── cmd/                     # Command line interface
│   ├── generate/            # Generate command implementation
│   │   └── generate.go      # Generate command (to be implemented)
│   ├── server/              # HTTP server command
│   │   └── server.go        # Server command (to be implemented)
│   ├── watch/               # Watch command for continuous monitoring
│   │   └── watch.go         # Watch command (to be implemented)
│   └── nomadmcppack.go      # Root command and CLI setup
├── pkg/                     # Public packages
│   ├── generate/            # Pack generation functionality
│   │   └── generate.go      # Generate Nomad MCP Server Packs functionality (to be implemented)
│   └── registry/            # MCP Registry client ✅ IMPLEMENTED
│       ├── registry.go      # HTTP client for MCP Registry API v1.0.0
│       └── registry_test.go # Comprehensive unit tests
├── internal/                # Private application code
│   ├── config/              # Configuration management ✅ IMPLEMENTED
│   │   └── config.go        # Viper-based configuration with env/file/defaults
│   ├── ctxutils/            # Context utilities
│   │   └── ctxutils.go      # Request-scoped logging, tracing, etc. (to be implemented)
│   ├── genutils/            # Generate command utilities
│   │   └── genutils.go      # Argument parsing, validation, etc. (to be implemented)
│   ├── logutils/            # Logging utilities
│   │   └── logutils.go      # Logger initialization, custom handler, etc. (to be implemented)
│   ├── serverutils/         # Server command utilities
│   │   └── serverutils.go   # HTTP server setup, routing, etc. (to be implemented)
│   ├── watch/               # Watch functionality
│   │   └── watch.go         # Registry polling, state management, etc. (to be implemented)
│   ├── watchutils/          # Watch command utilities
│   │   └── watchutils.go    # Watch-specific argument parsing, etc. (to be implemented)
│   └── templates/           # Embedded Nomad Pack templates
│       └── (to be implemented) # Go text/template definitions for pack generation
├── tests/                   # Test suite ✅ IMPLEMENTED
│   └── integration/         # Integration tests
│       ├── registry/        # Git submodule: MCP Registry (v1.0.0)
│       ├── registry_integration_test.go # Integration tests against local registry
│       ├── helpers_test.go  # Test utilities and registry detection
│       └── fixtures_test.go # Test data and sample server definitions
├── reference/               # Reference documentation ✅ IMPLEMENTED
│   └── registry-api-v1.0.0.yaml  # OpenAPI v1.0.0 specification for MCP Registry
├── main.go                  # Application entry point
├── go.mod                   # Go module definition
├── Makefile                 # Build automation with registry targets ✅ UPDATED
└── .gitmodules              # Git submodule configuration ✅ IMPLEMENTED
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
