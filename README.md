# nomad-mcp-pack

A command-line utility that generates HashiCorp Nomad Pack definitions from MCP (Model Context Protocol) servers registered in the official MCP Registry.

## Overview

`nomad-mcp-pack` bridges the gap between the MCP ecosystem and HashiCorp Nomad by automatically generating Nomad Pack templates from MCP server definitions. It interfaces with the official MCP Server Registry at [registry.modelcontextprotocol.io](https://registry.modelcontextprotocol.io) to fetch server metadata and create deployment-ready Nomad Packs.

## Features (Planned)

- **Automatic Pack Generation**: Generate Nomad Packs from MCP servers with supported package types (NPM, PyPI, OCI/Docker, NuGet)
- **Version Management**: Support for specific versions or latest version selection
- **Continuous Monitoring**: Watch mode for automatic pack generation when new servers are added or updated
- **HTTP API**: Server mode provides REST API endpoints for programmatic access
- **State Management**: Track generated packs to avoid duplicates in watch mode
- **Flexible Filtering**: Filter servers by package type and names in watch mode
- **Watch Mode Terminal UI Option**: Create a Terminal UI for operating in watch mode 

## Installation

### From Source

Requires Go 1.25.1 or later:

```bash
git clone https://github.com/leefowlercu/nomad-mcp-pack.git
cd nomad-mcp-pack
make install
```

This will build the binary and install it to `~/go/bin/`.

**Note**: The core functionality is currently under development. The CLI structure is in place but pack generation features are not yet implemented.

### Building Locally

```bash
make build
```

The binary will be created in the current directory as `nomad-mcp-pack`.

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
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest -o ./packs/astra-db

# Dry run to preview without creating files
nomad-mcp-pack generate io.github.modelcontextprotocol/filesystem@latest --dry-run

# Force overwrite existing pack
nomad-mcp-pack generate io.github.modelcontextprotocol/filesystem@1.0.2 --force

# Specify preferred package type when multiple are available
nomad-mcp-pack generate io.github.example/multi-package@latest --package-type oci
```

#### Watch Mode

Continuously monitor the registry for new or updated servers:

```bash
# Watch with default settings
nomad-mcp-pack watch

# Custom output directory and interval
nomad-mcp-pack watch --output ./packs --interval 300

# Watch with filtering by package type (only supported types)
nomad-mcp-pack watch --filter-package-type npm,oci

# Watch specific servers by name
nomad-mcp-pack watch --filter-names io.github.datastax/astra-db-mcp,io.github.example/server
```

#### HTTP Server Mode

Run as an HTTP server for programmatic access:

```bash
# Start on default port 8080
nomad-mcp-pack server

# Custom port
nomad-mcp-pack server --port 9090

# With TLS
nomad-mcp-pack server --tls-cert server.crt --tls-key server.key
```

### Global Options

- `--mcp-registry-addr`: Override the default MCP Registry URL (default: `https://registry.modelcontextprotocol.io`)
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
├── cmd/                     # Package `cmd`
│   ├── generate/            # Package `cmdgenerate`
│   │   └── generate.go      # Generate command
│   ├── server/              # Package `cmdserver`  
│   │   └── server.go        # Server command
│   ├── watch/               # Package `cmdwatch`
│   │   └── watch.go         # Watch command
│   └── nomadmcppack.go      # Nomad MCP Pack command
├── pkg/                     # Package `pkg` 
│   ├── generate/            # Package `generate` (to be implemented)
│   │   └── generate.go      # Generate Nomad MCP Server Packs functionality (to be implemented)
├── internal/                # Package `internal`
│   ├── config/              # Package `ctxutils`
│   │   └── config.go        # Internal Configuration utilities (Set Defaults, Read Config from File or Env)
│   ├── ctxutils/            # Package `ctxutils`
│   │   └── ctxutils.go      # Internal Context utilities (Request-Scoped Logging, Tracing, etc.)
│   ├── genutils/            # Package `genutils`
│   │   └── genutils.go      # Internal Generate command utilities (Argument Parsing, Validation, etc.)
│   ├── logutils/            # Package `logutils`
│   │   └── logutils.go      # Internal Logging utilities (Logger Initialization, Custom Handler, etc.)
│   ├── serverutils/         # Package `serverutils`
│   │   └── serverutils.go   # Internal Server command utilities (Argument Parsing, Validation, etc.)
│   ├── watch/               # Package `watch`
│   │   └── watch.go         # Internal Watch functionality (Polling Registry, Maintain Watch State, Use Generate functionality, etc.)
│   ├── watchutils/          # Package `watchutils`
│   │   └── watchutils.go    # Internal Watch command utilities (Argument Parsing, Validation, etc.)
│   └── templates/           # Embedded Nomad Pack templates (to be implemented)
├── main.go                  # Application entry point
├── go.mod                   # Go module definition
└── Makefile                 # Build automation
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
