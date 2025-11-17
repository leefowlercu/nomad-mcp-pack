# nomad-mcp-pack

Generate [HashiCorp Nomad Packs](https://github.com/hashicorp/nomad-pack) from [Model Context Protocol (MCP) Servers](https://modelcontextprotocol.io/) registered in the official MCP Registry.

## Features

- **Automatic Pack Generation**: Generate Nomad Packs from MCP Registry servers
- **Multiple Package Types**: Support for npm, pypi, OCI containers, and NuGet
- **Transport Protocols**: stdio, HTTP, and Server-Sent Events (SSE)
- **Continuous Monitoring**: Watch mode for automated pack updates
- **Dry-Run Mode**: Preview changes before execution
- **HTTP API Server** (Not Yet Implemented): Remote pack generation via REST API

## How It Works

nomad-mcp-pack automates the process of creating Nomad Pack definitions for MCP servers:

1. **Query MCP Registry**: Connects to the MCP Registry API to discover available servers
2. **Resolve Version**: Converts `@latest` syntax to specific semantic versions
3. **Validate Server**: Checks server status (active/deprecated/deleted) and availability
4. **Select Package & Transport**: Auto-detects or uses specified package type (npm/pypi/oci/nuget) and transport protocol (stdio/http/sse)
5. **Generate from Templates**: Renders Nomad job specifications using embedded templates with appropriate driver selection:
   - **Docker driver** for OCI container images
   - **Exec driver** for npm, pypi, and nuget packages with runtime installation
6. **Output Pack**: Creates pack directory or ZIP archive with metadata, variables, templates, and documentation

The generated packs are standard Nomad Pack definitions that can be deployed directly to Nomad clusters using `nomad-pack run`.

## Quick Start

```bash
# Install the tool
go install github.com/leefowlercu/nomad-mcp-pack@latest

# Check version
nomad-mcp-pack --version

# Generate a pack for an MCP server
nomad-mcp-pack generate com.falkordb/QueryWeaver@latest

# Watch for new/updated servers and auto-generate packs
nomad-mcp-pack watch --poll-interval 300

# Deploy a generated pack to Nomad cluster
# (Directory name is sanitized: slashes→dashes, dots→dashes)
cd packs/com-falkordb-QueryWeaver-<version>-<package>-<transport>
nomad-pack run .
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

### Global Flags

All commands support these global flags:

- `--silent, -s`: Suppress user-facing output (errors and warnings still shown)
- `--output-dir`: Output directory for generated packs (default: `./packs`)
- `--output-type`: Output type - `packdir` or `archive` (default: `packdir`)
- `--dry-run`: Show what would be done without making changes
- `--force-overwrite`: Overwrite existing pack directories/archives
- `--allow-deprecated`: Allow generation of packs for deprecated servers

### Generate Command

Generate a Nomad Pack for a specific MCP Server version:

```bash
# Generate pack for latest version
nomad-mcp-pack generate com.falkordb/QueryWeaver@latest

# Generate pack for specific version
nomad-mcp-pack generate com.falkordb/QueryWeaver@0.0.11

# Specify package type (npm, pypi, oci, nuget)
nomad-mcp-pack generate com.falkordb/QueryWeaver@latest --package-type oci

# Generate as ZIP archive instead of directory
nomad-mcp-pack generate com.falkordb/QueryWeaver@latest --output-type archive

# Dry run to preview without creating files
nomad-mcp-pack generate com.falkordb/QueryWeaver@latest --dry-run

# Force overwrite existing pack
nomad-mcp-pack generate com.falkordb/QueryWeaver@latest --force-overwrite

# Silent mode - suppress user-facing output (errors still shown)
nomad-mcp-pack generate com.falkordb/QueryWeaver@latest --silent
```

### Watch Command

Continuously monitor the MCP Registry and auto-generate packs:

```bash
# Watch all servers with default 5-minute polling
nomad-mcp-pack watch

# Custom poll interval (minimum 30 seconds)
nomad-mcp-pack watch --poll-interval 60

# Filter by exact server names
nomad-mcp-pack watch --filter-server-names "com.falkordb/QueryWeaver,io.github.pshivapr/selenium-mcp"

# Filter by package types
nomad-mcp-pack watch --filter-package-types "oci,npm"

# Control concurrency
nomad-mcp-pack watch --max-concurrent 10

# Custom state file location
nomad-mcp-pack watch --state-file ./my-watch-state.json

# Silent mode for automated environments
nomad-mcp-pack watch --silent --poll-interval 300
```

#### Watch State File Format

The watch command maintains a JSON state file (default: `./watch.json`) to track generated packs and prevent unnecessary regeneration. The state file is automatically created and updated as packs are generated.

**State File Structure:**

```json
{
  "last_poll": "2025-10-27T15:30:00Z",
  "servers": {
    "com.falkordb/QueryWeaver@0.0.11:oci:http": {
      "namespace": "com.falkordb",
      "name": "QueryWeaver",
      "version": "0.0.11",
      "package_type": "oci",
      "transport_type": "http",
      "updated_at": "2025-10-15T10:00:00Z",
      "generated_at": "2025-10-27T15:30:00Z",
      "checksum": ""
    }
  }
}
```

**Field Descriptions:**

- **`last_poll`**: Timestamp of the most recent registry poll
- **`servers`**: Map of server keys to their generation state
  - **Key format**: `namespace/name@version:package_type:transport_type`
  - **`namespace`**: MCP server namespace (e.g., `com.falkordb`)
  - **`name`**: MCP server name (e.g., `QueryWeaver`)
  - **`version`**: Server version that was generated
  - **`package_type`**: Package type used (`npm`, `pypi`, `oci`, `nuget`)
  - **`transport_type`**: Transport type used (`stdio`, `http`, `sse`)
  - **`updated_at`**: When the server was last updated in the registry
  - **`generated_at`**: When the pack was generated
  - **`checksum`**: Reserved for future use (currently empty)

**Regeneration Logic:**

The watch command regenerates a pack only when:
1. The server is not in the state file (new server)
2. The registry `updated_at` timestamp is newer than the local `generated_at` timestamp (server was updated)

**Inspecting State:**

```bash
# View formatted state file
cat watch.json | jq .

# Check which servers are tracked
cat watch.json | jq '.servers | keys'

# Find when a specific server was last generated
cat watch.json | jq '.servers["com.falkordb/QueryWeaver@0.0.11:oci:http"]'

# Check last poll time
cat watch.json | jq -r '.last_poll'
```

**State File Recovery:**

If the state file becomes corrupted or you want to force regeneration:

```bash
# Delete state file to regenerate all packs
rm ./watch.json

# Or use a different state file location
nomad-mcp-pack watch --state-file ./new-state.json
```

> **Note**: Deleting the state file will cause all matching servers to be regenerated on the next poll. Use `--dry-run` to preview what would be generated before committing to regeneration.

### Server Command (Not Yet Implemented)

> **Status**: This command is planned but not yet implemented. The command structure below shows the planned interface.

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

All configuration options can be set via environment variables with the `NOMAD_MCP_PACK_` prefix:

**Global Settings:**

| Variable | Description | Default |
|----------|-------------|---------|
| `NOMAD_MCP_PACK_REGISTRY_URL` | MCP Registry API URL | `https://registry.modelcontextprotocol.io/` |
| `NOMAD_MCP_PACK_LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `NOMAD_MCP_PACK_ENV` | Environment mode (dev, prod) | `dev` |
| `NOMAD_MCP_PACK_OUTPUT_DIR` | Pack output directory | `./packs` |
| `NOMAD_MCP_PACK_OUTPUT_TYPE` | Output format (packdir, archive) | `packdir` |
| `NOMAD_MCP_PACK_DRY_RUN` | Preview without creating files | `false` |
| `NOMAD_MCP_PACK_FORCE_OVERWRITE` | Overwrite existing packs | `false` |
| `NOMAD_MCP_PACK_ALLOW_DEPRECATED` | Include deprecated servers | `false` |
| `NOMAD_MCP_PACK_SILENT` | Suppress non-error output | `false` |

**Generate Command:**

| Variable | Description | Default |
|----------|-------------|---------|
| `NOMAD_MCP_PACK_GENERATE_PACKAGE_TYPE` | Default package type (npm, pypi, oci, nuget) | `""` (auto-detect) |
| `NOMAD_MCP_PACK_GENERATE_TRANSPORT_TYPE` | Default transport type (stdio, http, sse) | `""` (auto-detect) |

**Watch Command:**

| Variable | Description | Default |
|----------|-------------|---------|
| `NOMAD_MCP_PACK_WATCH_POLL_INTERVAL` | Poll interval in seconds (minimum 30) | `300` |
| `NOMAD_MCP_PACK_WATCH_FILTER_SERVER_NAMES` | Comma-separated server names to watch | `""` (all) |
| `NOMAD_MCP_PACK_WATCH_FILTER_PACKAGE_TYPES` | Comma-separated package types | `""` (all) |
| `NOMAD_MCP_PACK_WATCH_FILTER_TRANSPORT_TYPES` | Comma-separated transport types | `""` (all) |
| `NOMAD_MCP_PACK_WATCH_STATE_FILE` | State file location | `./watch.json` |
| `NOMAD_MCP_PACK_WATCH_MAX_CONCURRENT` | Max concurrent pack generations | `5` |

**Server Command** (Not Yet Implemented):

| Variable | Description | Default |
|----------|-------------|---------|
| `NOMAD_MCP_PACK_SERVER_ADDR` | Server bind address | `:8080` |
| `NOMAD_MCP_PACK_SERVER_READ_TIMEOUT` | Read timeout in seconds | `10` |
| `NOMAD_MCP_PACK_SERVER_WRITE_TIMEOUT` | Write timeout in seconds | `10` |

**Example:**

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
  filter_server_names: ["com.falkordb/QueryWeaver", "io.github.pshivapr/selenium-mcp"]
  filter_package_types: ["oci", "npm"]
  max_concurrent: 5

# Server configuration (feature not yet implemented)
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

> **Note**: The `http` transport type is internally mapped to `streamable-http` in the MCP Registry API.
> Both names refer to the same transport mechanism. You may see `streamable-http` in logs and pack names.

## Generated Pack Structure

Each generated pack includes:

```
<sanitized-server-name>-<version>-<package>-<transport>/
├── metadata.hcl           # Pack metadata and version info
├── variables.hcl          # Configurable variables
├── outputs.tpl            # Pack outputs template
├── README.md              # Generated documentation
└── templates/
    ├── mcp-server.nomad.tpl # Nomad job template
    └── _helpers.tpl         # Template helper functions
```

**Pack Naming**: Directory names are sanitized by converting:
- Forward slashes (`/`) → dashes (`-`)
- Dots (`.`) → dashes (`-`)
- Only alphanumeric characters, dashes, and underscores are retained

**Example**: `com.falkordb/QueryWeaver@0.0.11` with package `oci` and transport `http` becomes:
```
com-falkordb-QueryWeaver-0-0-11-oci-http/
```

### Example Pack Contents

**metadata.hcl**:
```hcl
app {
  name        = "QueryWeaver"
  version     = "0.0.11"
  description = "FalkorDB QueryWeaver MCP Server"
  url         = "https://registry.modelcontextprotocol.io/servers/com.falkordb/QueryWeaver"
}
```

**Job Template** (automatically selects appropriate Nomad driver):
- **OCI packages**: Uses `docker` driver
- **NPM/PyPI/NuGet packages**: Uses `exec` driver with package installation

### Service Name and Port Inference

For HTTP-based MCP servers (`http` and `sse` transports), the generator automatically infers service names and container ports:

**Service Name Inference:**

Service names are derived from the MCP server name by:
1. Taking the last component after the final slash
2. Converting to lowercase
3. Sanitizing (removing special characters)
4. Appending `-mcp` suffix

**Examples:**
- `com.falkordb/QueryWeaver` → `queryweaver-mcp`
- `io.github.myorg/my-server` → `my-server-mcp`
- `simple-server` → `simple-server-mcp`

**Container Port Inference:**

Default container ports are selected based on package type and runtime hints:

| Package Type | Runtime/Framework | Default Port |
|--------------|-------------------|--------------|
| `pypi` | Python | 5000 |
| `npm` | Node.js | 3000 |
| `oci` | Docker container | 8080 |
| `nuget` | .NET | 5000 |

These defaults can be overridden using the `container_port` variable when deploying with nomad-pack:

```bash
nomad-pack run my-pack --var="container_port=9000"
```

**Generated Variables:**

For HTTP-based servers, packs include these variables:
- `service_name` - Consul service registration name (default: inferred)
- `container_port` - Port the container listens on (default: inferred)
- `host_port` - Static host port allocation, 0 for dynamic (default: 0)
- `service_tags` - Additional tags for load balancer integration (default: [])
- `health_check_interval` - Health check frequency (default: 30s)
- `health_check_timeout` - Health check timeout (default: 5s)

## Common Workflows

### CI/CD Pack Generation

```bash
# Generate packs for specific servers
nomad-mcp-pack watch \
  --filter-server-names "io.github.myorg/server1,io.github.myorg/server2" \
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

### Live Deployment to Nomad

```bash
# Generate pack for Docker-based MCP server
nomad-mcp-pack generate com.falkordb/QueryWeaver@latest --package-type oci

# Deploy to Nomad cluster with static port
cd packs/com-falkordb-QueryWeaver-<version>-oci-streamable-http
nomad-pack run . \
  --var="host_port=8091" \
  --var='service_tags=["traefik.enable=true"]'

# Verify deployment
nomad job status com-falkordb-QueryWeaver-<version>-oci-streamable-http
```

## Limitations

Current known limitations of nomad-mcp-pack:

- **Watch Poll Interval Minimum**: The watch command enforces a minimum poll interval of 30 seconds to avoid overloading the MCP Registry.
- **No Wildcard Filter Support**: Server name filters in watch command require exact matches. Wildcards or regex patterns are not supported.
- **State File Location**: The watch command state file must be a local filesystem path. Remote storage (S3, etc.) is not supported.
- **Terminal UI Not Functional**: The `--enable-tui` flag exists in the watch command but the Terminal UI feature is not yet implemented.
- **HTTP API Server Not Implemented**: The `server` command is stubbed but not functional. See the [Server Command](#server-command-not-yet-implemented) section for details.
- **Pack Regeneration**: Changing pack generation settings requires manual deletion of existing packs when using `watch` mode, as the state file doesn't track configuration changes.

## Demos

The project includes comprehensive interactive demonstration scripts for the `generate`, `watch`, and end-to-end deployment workflows. These demos are ideal for presentations, training sessions, and understanding tool capabilities.

### Available Demos

#### [Generate Command Demo](./demo/generate/)

Demonstrates one-time pack generation for specific MCP servers.

```bash
cd demo/generate
./demo.sh         # Interactive mode
./demo.sh auto    # Automatic mode
```

**Features showcased:**
- Version resolution with @latest syntax
- Specific version targeting
- Package and transport type selection
- Output formats (directory/archive)
- Dry-run mode
- Force overwrite handling
- Error handling and validation

[View Generate Demo Documentation →](./demo/generate/README.md)

#### [Watch Command Demo](./demo/watch/)

Demonstrates continuous monitoring and automatic pack generation.

```bash
cd demo/watch
./demo.sh         # Interactive mode
./demo.sh auto    # Automatic mode
```

**Features showcased:**
- Continuous polling with configurable intervals
- State file persistence
- Server name filtering
- Package and transport type filtering
- Concurrent generation control
- Graceful shutdown handling
- Background process management

[View Watch Demo Documentation →](./demo/watch/README.md)

#### [Integration Demo](./demo/integration/)

Demonstrates complete end-to-end workflow from pack generation through live deployment to a Nomad cluster.

```bash
cd demo/integration
./demo.sh         # Interactive mode
./demo.sh auto    # Automatic mode
```

**Features showcased:**
- Live Nomad cluster deployment
- Docker-based MCP server deployment (QueryWeaver)
- Static port allocation for load balancer integration
- Automated authentication token generation
- Traefik routing configuration
- Claude Code MCP server integration
- Health checking and verification
- Complete cleanup workflow

**Note:** Requires a live Nomad cluster with proper access credentials.

[View Integration Demo Documentation →](./demo/integration/README.md)

### Prerequisites

**All Demos:**
- nomad-mcp-pack installed (`make install` from project root)
- Internet connection to access the MCP Registry

**Generate and Watch Demos:**
- Optional: `tree` (generate demo) and `jq` (watch demo) for enhanced output

**Integration Demo:**
- Live Nomad cluster with Docker driver enabled
- `nomad` and `nomad-pack` CLI tools installed
- Environment variables: `NOMAD_ADDR` and `NOMAD_TOKEN`
- `jq` for automated token generation
- Optional: Traefik load balancer for external access

See the [Demo README](./demo/README.md) for complete documentation, customization options, and troubleshooting.

## Troubleshooting

### Server Not Found in Registry

**Error**: `no server "server-name" found with version "X.X.X"`

**Causes**:
- Server name misspelled or doesn't exist in the registry
- Specific version doesn't exist for the server
- Server has been removed from the registry

**Solutions**:
```bash
# List all available servers (requires web search or registry browsing)
# Visit: https://registry.modelcontextprotocol.io/

# Try with @latest to get most recent version
nomad-mcp-pack generate server-name@latest

# Use --allow-deprecated to include deprecated servers
nomad-mcp-pack generate server-name@latest --allow-deprecated
```

### Package Type Not Available

**Error**: `package type "X" not found for server "server-name"`

**Cause**: The requested package type doesn't exist for this MCP server.

**Solutions**:
```bash
# Let the tool auto-detect available package types (omit --package-type)
nomad-mcp-pack generate server-name@latest

# Check the MCP Registry to see available package types for this server
# Visit: https://registry.modelcontextprotocol.io/servers/server-name
```

### Transport Type Not Available

**Error**: `transport type "X" not found for server "server-name" with package type "Y"`

**Cause**: The requested transport type doesn't exist for this server/package combination.

**Solutions**:
```bash
# Let the tool auto-detect available transport (omit --transport-type)
nomad-mcp-pack generate server-name@latest --package-type oci

# Check available transports in the MCP Registry
```

### Watch State File Corruption

**Error**: `failed to load state file` or `invalid JSON in state file`

**Solutions**:
```bash
# Delete the corrupted state file (watch will recreate it)
rm ./watch.json

# Or specify a new state file location
nomad-mcp-pack watch --state-file ./new-watch-state.json
```

### Registry Connection Failures

**Error**: `failed to connect to registry` or `timeout connecting to registry`

**Causes**:
- No internet connection
- Firewall blocking access to registry.modelcontextprotocol.io
- Registry temporarily unavailable

**Solutions**:
```bash
# Check internet connectivity
curl -I https://registry.modelcontextprotocol.io/

# Try with a longer timeout (if supported in future versions)
# Currently the timeout is not configurable

# Check firewall rules allow HTTPS to registry.modelcontextprotocol.io
```

### Pack Already Exists

**Error**: `pack directory already exists`

**Solutions**:
```bash
# Use --force-overwrite to replace existing pack
nomad-mcp-pack generate server-name@latest --force-overwrite

# Or manually delete the existing pack
rm -rf packs/server-name-version-package-transport/
```

### Invalid Poll Interval

**Error**: `poll interval must be at least 30 seconds`

**Cause**: Watch command poll interval set below the minimum of 30 seconds.

**Solution**:
```bash
# Use minimum 30 seconds or higher
nomad-mcp-pack watch --poll-interval 30
```

### Permission Denied Errors

**Error**: `permission denied` when creating packs

**Solutions**:
```bash
# Check output directory permissions
ls -la ./packs

# Specify a different output directory where you have write permissions
nomad-mcp-pack generate server-name@latest --output-dir ~/my-packs

# Or fix permissions on the packs directory
chmod 755 ./packs
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/leefowlercu/nomad-mcp-pack.git
cd nomad-mcp-pack

# Build binary
make build

# Install to ~/go/bin
make install
```

### Running Tests

```bash
# Run all tests (unit + integration)
make test

# Run only unit tests
make test-unit

# Run integration tests
make test-integration

# Integration tests with verbose output
make test-integration-verbose
```

### Local MCP Registry for Testing

For testing against a local MCP Registry instance:

```bash
# Initialize registry submodule (first time only)
make registry-init

# Start local registry (requires Docker)
make registry-up

# View registry logs
make registry-logs

# Stop local registry
make registry-down

# Update registry to latest version
make registry-update
```

The local registry runs on `http://localhost:8080` and includes seed data for testing.

### Build Configuration

- **Binary name**: `nomad-mcp-pack`
- **Version injection**: Set via ldflags at build time
- **Default version**: `1.0.0-alpha-0`
- **Install location**: `~/go/bin`

### Additional Make Targets

- `make clean` - Remove build artifacts and test output files
- `make rebuild` - Clean and rebuild from scratch

## Help & Support

- Run `nomad-mcp-pack --help` for command-line help
- Run `nomad-mcp-pack <command> --help` for command-specific help
- [Report issues](https://github.com/leefowlercu/nomad-mcp-pack/issues)
- [MCP Registry](https://registry.modelcontextprotocol.io/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [Nomad Pack Documentation](https://github.com/hashicorp/nomad-pack)
