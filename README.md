# nomad-mcp-pack

A command-line utility that generates HashiCorp Nomad Pack definitions from MCP (Model Context Protocol) servers registered in the official MCP Registry.

## Overview

`nomad-mcp-pack` bridges the gap between the MCP ecosystem and HashiCorp Nomad by automatically generating Nomad Pack templates from MCP server definitions. It interfaces with the official MCP Server Registry at [registry.modelcontextprotocol.io](https://registry.modelcontextprotocol.io) to fetch server metadata and create deployment-ready Nomad Packs.

## Features

- **Automatic Pack Generation**: Generate Nomad Packs from any MCP server in the registry
- **Multiple Package Types**: Support for Docker, NPM, PyPI, and GitHub package sources
- **Continuous Monitoring**: Watch mode for automatic pack generation when new servers are added
- **HTTP API**: Server mode provides REST API endpoints for programmatic access
- **Flexible Output**: Multiple output formats (text, JSON, YAML) for inspection and listing
- **State Management**: Track generated packs to avoid duplicates in watch mode

## Installation

### From Source

Requires Go 1.25.1 or later:

```bash
git clone https://github.com/leefowlercu/nomad-mcp-pack.git
cd nomad-mcp-pack
make install
```

This will build the binary and install it to `~/go/bin/`.

### Building Locally

```bash
make build
```

The binary will be created in the current directory as `nomad-mcp-pack`.

## Usage

### Basic Commands

#### List Available MCP Servers

```bash
# List all servers in table format
nomad-mcp-pack list

# Output as JSON
nomad-mcp-pack list --format json

# Filter by category
nomad-mcp-pack list --category "AI Tools"

# Filter by package type
nomad-mcp-pack list --package-type docker
```

#### Generate a Nomad Pack

```bash
# Generate pack for a specific server
nomad-mcp-pack generate mcp-github

# Specify output directory
nomad-mcp-pack generate mcp-github -o ./packs/github

# Dry run to preview without creating files
nomad-mcp-pack generate mcp-github --dry-run

# Force overwrite existing pack
nomad-mcp-pack generate mcp-github --force
```

#### Inspect Server Details

```bash
# View detailed server information
nomad-mcp-pack inspect mcp-github

# Output as JSON
nomad-mcp-pack inspect mcp-github --format json

# Show only package information
nomad-mcp-pack inspect mcp-github --packages
```

#### Watch Mode

Continuously monitor the registry for new or updated servers:

```bash
# Watch with default settings
nomad-mcp-pack watch

# Custom output directory and interval
nomad-mcp-pack watch --output ./packs --interval 300

# Watch with filtering
nomad-mcp-pack watch --category "AI Tools" --package-type docker
```

#### HTTP Server Mode

Run as an HTTP server for programmatic access:

```bash
# Start on default port 8080
nomad-mcp-pack server

# Custom port with CORS enabled
nomad-mcp-pack server --port 9090 --enable-cors

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
│   └── server.nomad.tpl  # Main Nomad job template
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
├── cmd/                   # CLI command definitions
│   ├── generate/          # Pack generation command
│   ├── inspect/           # Server inspection command
│   ├── list/              # Server listing command
│   ├── server/            # HTTP server command
│   └── watch/             # Watch mode command
├── internal/              # Private application code
│   └── (implementation)   # Command implementations
├── pkg/                   # Public library code
│   └── (libraries)        # Reusable packages
├── main.go                # Application entry point
├── Makefile               # Build automation
└── go.mod                 # Go module definition
```

### Contributing

This project is under active development. Key areas that need implementation:

- Registry client for MCP Server Registry API
- Pack generation engine with template processing
- HTTP API handlers for server mode
- State management for watch mode
- Configuration management system

See [CLAUDE.md](./CLAUDE.md) for detailed development guidelines and architecture notes.

## Configuration

### Environment Variables

- `NOMAD_MCP_PACK_REGISTRY_URL`: Override the default registry URL
- `NOMAD_MCP_PACK_OUTPUT_DIR`: Default output directory for generated packs
- `NOMAD_MCP_PACK_LOG_LEVEL`: Set logging level (debug, info, warn, error)

### Multi-Tenant Support

Generated packs include variables for multi-tenant deployments:

- `tenant_id`: Namespace isolation
- `deployment_id`: Unique deployment identifier
- `user_id`: Track deployment ownership

## Limitations

- Currently does not support remote MCP servers (servers without package definitions)
- Requires packages to be publicly accessible
- Assumes HTTP transport for multi-tenant deployments

## License

[License information to be added]

## Related Projects

- [MCP Registry](https://registry.modelcontextprotocol.io) - Official MCP Server Registry
- [HashiCorp Nomad](https://www.nomadproject.io) - Simple and flexible workload orchestrator
- [Nomad Pack](https://github.com/hashicorp/nomad-pack) - Package manager for Nomad
- [Model Context Protocol](https://modelcontextprotocol.io) - Open protocol for AI model context