# nomad-mcp-pack

Generate [HashiCorp Nomad Packs](https://github.com/hashicorp/nomad-pack) from [Model Context Protocol (MCP) Servers](https://modelcontextprotocol.io/) registered in the official MCP Registry.

## Quick Start

```bash
# Install the tool
go install github.com/leefowlercu/nomad-mcp-pack@latest

# Generate a pack for an MCP server
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest

# Watch for new/updated servers and auto-generate packs
nomad-mcp-pack watch --poll-interval 300

# Start HTTP API server
nomad-mcp-pack server --addr :8080
```

## Installation

### Prerequisites

- Go 1.25+ (for building from source)
- Internet connection to access the [MCP Registry](https://registry.modelcontextprotocol.io)

### Install Binary

```bash
go install github.com/leefowlercu/nomad-mcp-pack@latest
```

### Build from Source

```bash
git clone https://github.com/leefowlercu/nomad-mcp-pack.git
cd nomad-mcp-pack
make install
```

## Usage

### Generate Command

Generate a Nomad Pack for a specific MCP Server version:

```bash
# Generate pack for latest version
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest

# Generate pack for specific version
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@0.0.1-seed

# Specify package type (npm, pypi, oci, nuget)
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --package-type oci

# Generate as ZIP archive instead of directory
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --output-type archive

# Dry run to preview without creating files
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --dry-run

# Force overwrite existing pack
nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --force-overwrite
```

### Watch Command

Continuously monitor the MCP Registry and auto-generate packs:

```bash
# Watch all servers with default 5-minute polling
nomad-mcp-pack watch

# Custom poll interval (minimum 30 seconds)
nomad-mcp-pack watch --poll-interval 60

# Filter by exact server names
nomad-mcp-pack watch --filter-names "io.github.fastapi/fastapi-mcp,io.github.example/sql-server"

# Filter by package types
nomad-mcp-pack watch --filter-package-types "oci,npm"

# Control concurrency
nomad-mcp-pack watch --max-concurrent 10

# Custom state file location
nomad-mcp-pack watch --state-file ./my-watch-state.json
```

### Server Command

Start an HTTP API server for remote pack generation:

```bash
# Start server on default port (:8080)
nomad-mcp-pack server

# Custom port and timeouts
nomad-mcp-pack server --addr :9090 --read-timeout 30 --write-timeout 30
```

## Configuration

### Configuration Hierarchy

Configuration is loaded in order of precedence (highest to lowest):

1. **Command-line flags**
2. **Environment variables** (`NOMAD_MCP_PACK_*`)
3. **Configuration file** (`config.yaml`)
4. **Built-in defaults**

### Environment Variables

```bash
export NOMAD_MCP_PACK_REGISTRY_URL="https://registry.modelcontextprotocol.io/"
export NOMAD_MCP_PACK_OUTPUT_DIR="./my-packs"
export NOMAD_MCP_PACK_LOG_LEVEL="debug"
export NOMAD_MCP_PACK_WATCH_POLL_INTERVAL="600"
```

### Configuration File

Place `config.yaml` in current directory or `~/.nomad-mcp-pack/`:

```yaml
registry_url: https://registry.modelcontextprotocol.io/
output_dir: ./packs
log_level: info
env: prod

watch:
  poll_interval: 300
  filter_names: ["io.github.fastapi/fastapi-mcp", "io.github.example/database-server"]
  filter_package_types: ["oci", "npm"]
  max_concurrent: 5

server:
  addr: :8080
  read_timeout: 10
  write_timeout: 10
```

## Package Types

| Type | Description | Driver |
|------|-------------|---------|
| `npm` | Node.js packages from npmjs.com | `exec` |
| `pypi` | Python packages from pypi.org | `exec` |
| `oci` | Container images from Docker registries | `docker` |
| `nuget` | .NET packages from nuget.org | `exec` |

## Transport Types

| Type | Description |
|------|-------------|
| `stdio` | Standard input/output communication |
| `http` | HTTP-based communication |
| `sse` | Server-Sent Events communication |

## Generated Pack Structure

Each generated pack includes:

```
mcp-server-name-version-package-transport/
├── metadata.hcl           # Pack metadata and version info
├── variables.hcl          # Configurable variables
├── outputs.tpl            # Pack outputs template
├── README.md              # Generated documentation
└── templates/
    └── mcp-server.nomad.tpl # Nomad job template
```

### Example Pack Contents

**metadata.hcl**:
```hcl
app {
  name        = "astra-db-mcp"
  version     = "0.0.1-seed"
  description = "DataStax Astra DB MCP Server"
  url         = "https://registry.modelcontextprotocol.io/servers/io.github.datastax/astra-db-mcp"
}
```

**Job Template** (automatically selects appropriate Nomad driver):
- **OCI packages**: Uses `docker` driver
- **NPM/PyPI/NuGet packages**: Uses `exec` driver with package installation

## Common Workflows

### CI/CD Pack Generation

```bash
# Generate packs for specific servers
nomad-mcp-pack watch \
  --filter-names "io.github.myorg/server1,io.github.myorg/server2" \
  --output-type archive \
  --poll-interval 300 \
  --max-concurrent 10
```

### Development Testing

```bash
# Test pack generation without creating files
nomad-mcp-pack generate myserver@latest --dry-run

# Generate to temporary directory
nomad-mcp-pack generate myserver@latest --output-dir /tmp/test-packs
```

### Production Monitoring

```bash
# Monitor with structured logging
NOMAD_MCP_PACK_ENV=prod \
NOMAD_MCP_PACK_LOG_LEVEL=error \
nomad-mcp-pack watch \
  --filter-package-types "oci" \
  --poll-interval 600 \
  --max-concurrent 5
```

## Help & Support

- Run `nomad-mcp-pack --help` for command-line help
- Run `nomad-mcp-pack <command> --help` for command-specific help
- [Report issues](https://github.com/leefowlercu/nomad-mcp-pack/issues)
- [MCP Registry](https://registry.modelcontextprotocol.io/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [Nomad Pack Documentation](https://github.com/hashicorp/nomad-pack)
